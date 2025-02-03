package main

import (
	"device-chronicle-server/cmd"
)

//
//func main() {
//	r := gin.Default()
//	wsServer := NewWebSocketServer()
//
//	// WebSocket endpoint
//	r.GET("/ws", wsServer.handleClient)
//	r.GET("/clients", wsServer.listClients)
//	// Run server
//	r.Run(":8080")
//}

func main() {
	cmd.Execute()
}
