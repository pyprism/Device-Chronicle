package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

//func TestHandleClient(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//	router := gin.Default()
//	wsServer := NewWebSocketServer()
//	router.GET("/ws", wsServer.HandleClient)
//
//	w := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/ws?client_id=test-client", nil)
//	router.ServeHTTP(w, req)
//
//	if w.Code != http.StatusSwitchingProtocols {
//		log.Printf("Error TestHandleClient: %v", w.Body.String())
//	}
//	assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
//}

func TestServeAnalyticsPage(t *testing.T) {
	router := gin.Default()
	router.LoadHTMLGlob("../templates/*")
	wsServer := NewWebSocketServer()
	router.GET("/analytics/:client_id", wsServer.ServeAnalyticsPage)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/analytics/test-client", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		log.Printf("Error TestServeAnalyticsPage: %v", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

//func TestHandleAnalytics(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//	router := gin.Default()
//	wsServer := NewWebSocketServer()
//	router.GET("/analytics/ws/:client_id", wsServer.HandleAnalytics)
//
//	w := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/analytics/ws/test-client", nil)
//	router.ServeHTTP(w, req)
//
//	if w.Code != http.StatusSwitchingProtocols {
//		log.Printf("Error TestHandleAnalytics: %v", w.Body.String())
//	}
//	assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
//}

//func TestListClients(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//	router := gin.Default()
//	wsServer := NewWebSocketServer()
//	router.GET("/clients", wsServer.ListClients)
//
//	w := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/clients", nil)
//	router.ServeHTTP(w, req)
//
//	if w.Code != http.StatusOK {
//		log.Printf("Error TestListClients: %v", w.Body.String())
//	}
//	assert.Equal(t, http.StatusOK, w.Code)
//}
