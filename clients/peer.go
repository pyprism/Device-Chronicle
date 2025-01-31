package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

func main() {
	// Generate a unique client ID
	clientID := uuid.New().String() // Example: "550e8400-e29b-41d4-a716-446655440000"
	fmt.Println("Generated Client ID:", clientID)

	// Connect to WebSocket signaling server
	signalingURL := fmt.Sprintf("ws://localhost:8080/signal?client_id=%s", clientID)
	conn, _, err := websocket.DefaultDialer.Dial(signalingURL, nil)
	if err != nil {
		log.Fatal("Failed to connect to signaling server:", err)
	}
	defer conn.Close()

	// Send a test message via WebSocket to check connection
	testMessage := fmt.Sprintf(`{"type":"test","message":"Hello from Golang Client %s"}`, clientID)
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
	if err != nil {
		log.Fatal("Failed to send test message:", err)
	}
	fmt.Println("Sent test message to server")

	// Listen for messages from the signaling server
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading from signaling server:", err)
				return
			}
			fmt.Printf("Received from server: %s\n", string(message))
		}
	}()

	// Create WebRTC peer connection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatal("Failed to create peer connection:", err)
	}

	// Create a WebRTC data channel
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		log.Fatal("Failed to create data channel:", err)
	}

	// When data channel opens, send a message
	dataChannel.OnOpen(func() {
		fmt.Println("Data channel open! Sending WebRTC message...")
		dataChannel.SendText("Hello from Golang Client via WebRTC!")
	})

	// Handle WebRTC messages
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Received WebRTC message: %s\n", string(msg.Data))
	})

	// Handle ICE candidates
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			candidateJSON, _ := json.Marshal(candidate.ToJSON())
			conn.WriteMessage(websocket.TextMessage, candidateJSON)
		}
	})

	// Create an offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatal("Failed to create offer:", err)
	}
	peerConnection.SetLocalDescription(offer)

	// Send offer to signaling server
	offerJSON, _ := json.Marshal(offer)
	conn.WriteMessage(websocket.TextMessage, offerJSON)

	// Keep running to receive messages
	select {}
}
