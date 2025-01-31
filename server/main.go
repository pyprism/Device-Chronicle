package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Store active clients
type WebSocketServer struct {
	clients map[string]*websocket.Conn
	mu      sync.Mutex
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients: make(map[string]*websocket.Conn),
	}
}

// WebSocket handler for clients
func (s *WebSocketServer) handleClient(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	// Upgrade connection
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

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", clientID)
			break
		}
		fmt.Printf("Received from %s: %s\n", clientID, string(msg))

		// Broadcast message
		s.broadcast(clientID, msg)
	}

	// Remove client when disconnected
	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()
}

// Broadcast message to all connected clients
func (s *WebSocketServer) broadcast(senderID string, msg []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientID, conn := range s.clients {
		if clientID != senderID { // Don't send back to the sender
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Printf("Error sending to %s: %v\n", clientID, err)
			}
		}
	}
}

// **New API to list all connected clients**
func (s *WebSocketServer) listClients(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientIDs := []string{}
	for clientID := range s.clients {
		clientIDs = append(clientIDs, clientID)
	}

	c.JSON(http.StatusOK, gin.H{"clients": clientIDs})
}

func main() {
	r := gin.Default()
	wsServer := NewWebSocketServer()

	// WebSocket endpoint
	r.GET("/ws", wsServer.handleClient)
	r.GET("/clients", wsServer.listClients)
	// Run server
	r.Run(":8080")
}
