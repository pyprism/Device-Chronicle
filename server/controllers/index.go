package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ServeIndexPage serves the main index page with client list
func (s *WebSocketServer) ServeIndexPage(c *gin.Context) {
	// Get list of all connected clients
	s.mu.RLock()
	clientIDs := make([]string, 0, len(s.clients))
	for client := range s.clients {
		clientIDs = append(clientIDs, client)
	}
	s.mu.RUnlock()

	c.HTML(http.StatusOK, "index.html", gin.H{
		"clients": clientIDs,
	})
}
