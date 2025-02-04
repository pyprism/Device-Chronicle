package main

import (
	"device-chronicle-client/utils"
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

	//data := make(map[string]interface{})
	//var err error
	//
	//if runtime.GOOS == "linux" {
	//	data, err = os.Linux()
	//}
	//if err != nil {
	//	log.Println("Error getting data:", err)
	//	return
	//}

	utils.Websocket(serverAddr)
}
