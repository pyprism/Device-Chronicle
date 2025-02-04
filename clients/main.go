package main

import (
	"device-chronicle-client/websocket"
	"flag"
	"log"
)

func main() {
	serverAddr := flag.String("server", "", "Server address, e.g. ws://localhost:8000")
	//interval := flag.Int("interval", 2, "Interval in seconds to send data to server")
	flag.Parse()

	if *serverAddr == "" {
		log.Fatalln("Server address is required. Usage: ./chronicle --server ws://localhost:8000")
	}

	websocket.Websocket(serverAddr)
}
