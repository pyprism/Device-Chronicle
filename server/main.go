package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SignalServer struct {
	clients map[string]*websocket.Conn
	mu      sync.Mutex
}

func NewSignalServer() *SignalServer {
	return &SignalServer{
		clients: make(map[string]*websocket.Conn),
	}
}

func (s *SignalServer) handleSignaling(c *gin.Context) {
	clientID := c.Query("client_id") // Get client_id from query
	if clientID == "" {
		fmt.Println("Error: client_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	fmt.Println("Client connecting:", clientID)

	// Upgrade to WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to upgrade WebSocket:", err)
		return
	}
	defer conn.Close()

	// Store client connection
	s.mu.Lock()
	s.clients[clientID] = conn
	s.mu.Unlock()

	fmt.Println("Client connected:", clientID)

	// Listen for messages from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", clientID)
			break
		}
		fmt.Printf("Received from %s: %s\n", clientID, string(message))

		// Relay messages if needed
		s.broadcastMessage(clientID, message)
	}

	// Remove client when they disconnect
	s.mu.Lock()
	delete(s.clients, clientID)
	s.mu.Unlock()
}

func (s *SignalServer) broadcastMessage(senderID string, message []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientID, conn := range s.clients {
		if clientID != senderID {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Printf("Failed to send message to %s: %v\n", clientID, err)
			}
		}
	}
}

func main() {
	r := gin.Default()
	signalServer := NewSignalServer()

	// Accept WebSocket connections
	r.GET("/signal", signalServer.handleSignaling)

	r.Run(":8080")
}
