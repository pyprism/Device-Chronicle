package main

import (
	"device-chronicle-client/websocket"
	"flag"
	"log"
)

func main() {
	serverAddr := flag.String("server", "", "Server address, e.g. http://localhost:8000")
	dummyData := flag.Bool("dummy", false, "Use dummy data instead of real data for testing")
	interval := flag.Int("interval", 2, "Interval in seconds to send data to server")
	clientName := flag.String("client", "", "Client name")
	flag.Parse()

	if *serverAddr == "" {
		log.Fatalln("Server address is required. Usage: ./chronicle --server http://localhost:8000")
	}

	if *clientName == "" {
		log.Fatalln("Client name is required. Usage: ./chronicle --client hiren's-desktop")
	}

	websocket.Websocket(serverAddr, dummyData, interval, clientName)
}
