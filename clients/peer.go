package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func main() {
	// Generate a unique client ID
	clientID := uuid.New().String()
	fmt.Println("Generated Client ID:", clientID)

	// Connect to WebSocket server
	serverURL := fmt.Sprintf("ws://localhost:8000/ws?client_id=%s", clientID)
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Failed to connect to WebSocket server:", err)
	}
	defer conn.Close()

	// Send data every 5 seconds
	for {
		message := fmt.Sprintf(`{"client_id": "%s", "temperature": 25, "humidity": 60}`, clientID)
		//message := fmt.Sprintf("Hello from Client %s", clientID)
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Write error:", err)
			return
		}
		fmt.Println("Sent:", message)

		time.Sleep(5 * time.Second)
	}
}
