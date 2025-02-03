package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"sync"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketServer Store active clients and WebSocket connections for analytics
type WebSocketServer struct {
	clients       map[string]*websocket.Conn
	analyticsConn map[string]*websocket.Conn // Store WebSocket connections for analytics
	mu            sync.Mutex
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:       make(map[string]*websocket.Conn),
		analyticsConn: make(map[string]*websocket.Conn),
	}
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
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Store connection
	s.mu.Lock()
	s.clients[clientID] = conn
	s.mu.Unlock()

	fmt.Println("Client connected:", clientID)

	// Read messages from client
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", clientID)
			break
		}
		fmt.Printf("Received from %s: %s\n", clientID, string(msg))

		// Forward message to analytics WebSocket if connected
		s.mu.Lock()
		if analyticsConn, ok := s.analyticsConn[clientID]; ok {
			analyticsConn.WriteMessage(websocket.TextMessage, msg)
		}
		s.mu.Unlock()
	}

	// Remove client when disconnected
	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()
}

// Serve analytics HTML page
func (s *WebSocketServer) ServeAnalyticsPage(c *gin.Context) {
	clientID := c.Param("client_id")
	tmpl, err := template.ParseFiles("templates/analytics.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading template")
		return
	}
	tmpl.Execute(c.Writer, gin.H{"client_id": clientID})
}

// Handle WebSocket for analytics
func (s *WebSocketServer) HandleAnalytics(c *gin.Context) {
	clientID := c.Param("client_id")

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Store connection for analytics
	s.mu.Lock()
	s.analyticsConn[clientID] = conn
	s.mu.Unlock()

	fmt.Println("Analytics client connected for:", clientID)

	// Keep the connection open
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Analytics disconnected:", clientID)
			break
		}
	}

	// Remove connection on disconnect
	s.mu.Lock()
	delete(s.analyticsConn, clientID)
	s.mu.Unlock()
}

// API to list all connected clients
func (s *WebSocketServer) ListClients(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientIDs := []string{}
	for clientID := range s.clients {
		clientIDs = append(clientIDs, clientID)
	}

	c.JSON(http.StatusOK, gin.H{"clients": clientIDs})
}
