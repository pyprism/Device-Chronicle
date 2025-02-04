package utils

import (
	data2 "device-chronicle-client/data"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

////func Websocket(serverAddr *string) {
//	// Generate a unique client ID
//	clientID := RandStringBytes(8)
//	log.Println("Sending data to WebSocket server with Client ID:", clientID)
//
//	// Connect to WebSocket server with retry mechanism
//	serverURL := fmt.Sprintf("%s/ws?client_id=%s", *serverAddr, clientID)
//	var conn *data.Conn
//	var err error
//	retryInterval := 5 * time.Second
//
//	for {
//		conn, _, err = data.DefaultDialer.Dial(serverURL, nil)
//		if err != nil {
//			log.Println("Failed to connect to WebSocket server:", err)
//			log.Printf("Retrying in %v...\n", retryInterval)
//			time.Sleep(retryInterval)
//			continue
//		}
//		break
//	}
//	defer conn.Close()
//
//	ticker := time.NewTicker(time.Duration(2) * time.Second)
//	defer ticker.Stop()
//	// Send data every 5 seconds
//	for {
//		select {
//		case <-ticker.C:
//			data := make(map[string]interface{})
//			var err error
//			if runtime.GOOS == "linux" {
//				data, err = os.Linux()
//			}
//			if err != nil {
//				log.Println("Error getting data:", err)
//				continue
//			}
//			message, err := json.Marshal(data)
//			if err != nil {
//				log.Println("Error marshalling data:", err)
//				return
//			}
//			err = conn.WriteMessage(data.TextMessage, []byte(message))
//			if err != nil {
//				log.Println("Write error:", err)
//				return
//			}
//		}
//
//	}
//}

func Websocket(serverAddr *string) {
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
			data, err := data2.FetchData()
			if err != nil {
				log.Println("Error getting data:", err)
				continue
			}

			if err := sendData(conn, data); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}

func generateClientID() string {
	return RandStringBytes(8)
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

//func fetchData() (map[string]interface{}, error) {
//	if runtime.GOOS == "linux" {
//		return os.Linux()
//	}
//	return nil, fmt.Errorf("unsupported OS")
//}

func sendData(conn *websocket.Conn, data map[string]interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}
	return conn.WriteMessage(websocket.TextMessage, message)
}
