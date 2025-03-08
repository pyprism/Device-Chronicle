package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type Option func(*WebSocketServer)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketServer Store active clients and WebSocket connections for analytics
type WebSocketServer struct {
	clients       map[string][]*websocket.Conn
	analyticsConn map[string][]*websocket.Conn
	mu            sync.RWMutex // Add mutex for thread safety
	logger        *zap.Logger
}

func WithLogger(logger *zap.Logger) Option {
	return func(ws *WebSocketServer) {
		ws.logger = logger
	}
}

func NewWebSocketServer(opts ...Option) *WebSocketServer {
	ws := &WebSocketServer{
		clients:       make(map[string][]*websocket.Conn),
		analyticsConn: make(map[string][]*websocket.Conn),
	}

	// Apply options
	for _, opt := range opts {
		opt(ws)
	}

	// Set default logger if none provided
	if ws.logger == nil {
		ws.logger, _ = zap.NewProduction()
	}

	return ws
}

// HandleClient Handle client WebSocket connection on /ws
func (s *WebSocketServer) HandleClient(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed:", zap.Error(err))
		return
	}
	defer conn.Close()

	// Store connection
	s.mu.Lock()
	s.clients[clientID] = append(s.clients[clientID], conn)
	s.mu.Unlock()

	s.logger.Info("Client connected", zap.String("clientID", clientID))

	// Read messages from client
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.logger.Info("Client disconnected", zap.String("clientID", clientID), zap.Error(err))
			break
		}
		//s.logger.Info("Received from client", zap.String("clientID", clientID), zap.String("message", string(msg)))

		// Forward message to analytics WebSocket if connected
		s.mu.RLock()
		if analyticsConns, ok := s.analyticsConn[clientID]; ok {
			for _, analyticsConn := range analyticsConns {
				if err := analyticsConn.WriteMessage(websocket.TextMessage, msg); err != nil {
					s.logger.Error("Failed to forward message to analytics", zap.String("clientID", clientID), zap.Error(err))
				}
			}
		}
		s.mu.RUnlock()
	}

	// Remove client when disconnected
	s.mu.Lock()
	connections := s.clients[clientID]
	for i, c := range connections {
		if c == conn { // Fixed variable shadowing bug
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	if len(connections) == 0 {
		delete(s.clients, clientID)
	} else {
		s.clients[clientID] = connections // Update the connections list
	}
	s.mu.Unlock()
}

// ServeAnalyticsPage Serve analytics HTML page
func (s *WebSocketServer) ServeAnalyticsPage(c *gin.Context) {
	clientID := c.Param("client_id")

	// Get list of all connected clients for the dropdown
	s.mu.RLock()
	clientIDs := make([]string, 0, len(s.clients))
	for client := range s.clients {
		clientIDs = append(clientIDs, client)
	}
	s.mu.RUnlock()

	version := time.Now().Unix()

	c.HTML(http.StatusOK, "analytics.html", gin.H{
		"client_id": clientID,
		"clients":   clientIDs,
		"version":   version, // to bypass cdn cache for custom.js
	})
}

// HandleAnalytics Handle WebSocket for analytics
func (s *WebSocketServer) HandleAnalytics(c *gin.Context) {
	clientID := c.Param("client_id")

	// Check if client is connected
	s.mu.RLock()
	_, exists := s.clients[clientID]
	s.mu.RUnlock()

	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Only upgrade to WebSocket if client exists
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Store connection for analytics
	s.mu.Lock()
	s.analyticsConn[clientID] = append(s.analyticsConn[clientID], conn)
	s.mu.Unlock()

	s.logger.Info("Analytics client connected", zap.String("clientID", clientID))

	// Keep the connection open
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.logger.Info("Analytics disconnected", zap.String("clientID", clientID), zap.Error(err))
			break
		}
	}

	// Remove connection on disconnect
	s.mu.Lock()
	connections := s.analyticsConn[clientID]
	for i, c := range connections {
		if c == conn {
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// If there are no more connections for this client, remove it from the map
	if len(connections) == 0 {
		delete(s.analyticsConn, clientID)
	} else {
		s.analyticsConn[clientID] = connections
	}
	s.mu.Unlock()
}

// ListClients API to list all connected clients
func (s *WebSocketServer) ListClients(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientIDs := []string{}
	for clientID := range s.clients {
		clientIDs = append(clientIDs, clientID)
	}

	c.JSON(http.StatusOK, gin.H{"clients": clientIDs})
}
