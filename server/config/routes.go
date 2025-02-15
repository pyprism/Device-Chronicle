package config

import (
	"device-chronicle-server/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	wsServer := controllers.NewWebSocketServer()
	staticFile := controllers.StaticFileController{}
	router.GET("/ws", wsServer.HandleClient)
	router.GET("/analytics/:client_id", wsServer.ServeAnalyticsPage)
	router.GET("/analytics_ws/:client_id", wsServer.HandleAnalytics)
	router.GET("/clients", wsServer.ListClients)
	router.GET("/:static", staticFile.StaticFile)
}
