package websocket

import (
	"device-chronicle-client/fetch"
	"device-chronicle-client/models"
	"device-chronicle-client/utils"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"
)

func Websocket(serverAddr *string, dummy *bool, interval *int, clientName *string) {
	clientID := *clientName
	log.Println("Sending data to WebSocket server with Client ID:", clientID)

	conn, err := connectToServer(serverAddr, clientID)
	if err != nil {
		log.Println("Failed to connect to WebSocket server:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			var systemData *models.System
			var err error

			if *dummy {
				systemData = utils.DummyData()
			} else {
				systemData, err = fetch.FetchData()
				if err != nil {
					log.Println("Error getting data:", err)
					continue
				}
			}

			dataMap := systemData.ToMap()

			// if there is an error sending data, close the connection and reconnect
			if err := sendData(conn, dataMap); err != nil {
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
	protocol := "ws"
	if strings.HasPrefix(*serverAddr, "https") {
		protocol = "wss"
	}

	// Remove http/https from serverAddr
	trimmedAddr := strings.TrimPrefix(*serverAddr, "http://")
	trimmedAddr = strings.TrimPrefix(trimmedAddr, "https://")

	serverURL := fmt.Sprintf("%s://%s/ws?client_id=%s", protocol, trimmedAddr, clientID)
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
