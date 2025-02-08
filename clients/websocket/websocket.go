package websocket

import (
	fetch "device-chronicle-client/data"
	"device-chronicle-client/utils"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func Websocket(serverAddr *string, dummy *bool) {
	clientID := generateClientID()
	log.Println("Sending data to WebSocket server with Client ID:", clientID)

	conn, err := connectToServer(serverAddr, clientID)
	if err != nil {
		log.Println("Failed to connect to WebSocket server:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := map[string]interface{}{}, error(nil)
			if *dummy {
				data = utils.DummyData()
			} else {
				data, err = fetch.FetchData()
				if err != nil {
					log.Println("Error getting data:", err)
					continue
				}
			}

			// if there is an error sending data, close the connection and reconnect
			if err := sendData(conn, data); err != nil {
				log.Println("Write error:", err)
				conn.Close()
				conn, err = connectToServer(serverAddr, clientID)
				if err != nil {
					log.Println("Failed to reconnect to WebSocket server:", err)
					return
				}
			}
		}
	}
}

func generateClientID() string {
	return utils.RandStringBytes(8)
}

func connectToServer(serverAddr *string, clientID string) (*websocket.Conn, error) {
	serverURL := fmt.Sprintf("%s/ws?client_id=%s", *serverAddr, clientID)
	retryInterval := 5 * time.Second

	for {
		conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
		if err != nil {
			log.Println("Failed to connect to WebSocket server:", err)
			log.Printf("Retrying in %v...\n", retryInterval)
			time.Sleep(retryInterval)
			continue
		}
		return conn, nil
	}
}

func sendData(conn *websocket.Conn, data map[string]interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}
	return conn.WriteMessage(websocket.TextMessage, message)
}
