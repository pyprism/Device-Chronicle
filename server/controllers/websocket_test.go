package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testSetup struct {
	router   *gin.Engine
	wsServer *WebSocketServer
	server   *httptest.Server
}

func setupTest() *testSetup {
	router := gin.Default()
	wsServer := NewWebSocketServer()
	return &testSetup{
		router:   router,
		wsServer: wsServer,
	}
}

func setupWebSocketServer(ts *testSetup) {
	ts.server = httptest.NewServer(ts.router)
}

func TestHandleClient(t *testing.T) {
	ts := setupTest()
	ts.router.GET("/ws", ts.wsServer.HandleClient)
	setupWebSocketServer(ts)
	defer ts.server.Close()

	tests := []struct {
		name     string
		clientID string
		wantErr  bool
	}{
		{"valid client", "test-client", false},
		{"empty client id", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsURL := "ws" + strings.TrimPrefix(ts.server.URL, "http") + "/ws"
			if tt.clientID != "" {
				wsURL += "?client_id=" + tt.clientID
			}

			ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			defer ws.Close()

			assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

			// Test message sending
			message := []byte("test message")
			err = ws.WriteMessage(websocket.TextMessage, message)
			require.NoError(t, err)

			// Verify client is registered
			assert.Contains(t, ts.wsServer.clients, tt.clientID)
		})
	}
}

func TestServeAnalyticsPage(t *testing.T) {
	ts := setupTest()
	ts.router.LoadHTMLGlob("../templates/*")
	ts.router.GET("/analytics/:client_id", ts.wsServer.ServeAnalyticsPage)

	tests := []struct {
		name     string
		clientID string
		wantCode int
	}{
		{"valid client", "test-client", http.StatusOK},
		{"empty client id", "", http.StatusNotFound}, // Changed expected status
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/analytics/"
			if tt.clientID != "" {
				path += tt.clientID
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", path, nil)
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				assert.Contains(t, w.Body.String(), "Device Chronicle")
			}
		})
	}
}

func TestHandleAnalytics(t *testing.T) {
	ts := setupTest()
	ts.router.GET("/analytics/ws/:client_id", ts.wsServer.HandleAnalytics)
	setupWebSocketServer(ts)
	defer ts.server.Close()

	tests := []struct {
		name     string
		clientID string
	}{
		{"valid analytics connection", "test-client"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsURL := "ws" + strings.TrimPrefix(ts.server.URL, "http") + "/analytics/ws/" + tt.clientID

			ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			defer ws.Close()

			assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

			// Verify analytics connection is registered
			time.Sleep(100 * time.Millisecond) // Allow time for connection to be registered
			assert.Contains(t, ts.wsServer.analyticsConn, tt.clientID)
		})
	}
}

func TestListClients(t *testing.T) {
	ts := setupTest()
	ts.router.GET("/clients", ts.wsServer.ListClients)

	tests := []struct {
		name         string
		setupClients bool
		expectedJSON string
		expectedCode int
	}{
		{
			name:         "empty clients list",
			setupClients: false,
			expectedJSON: `{"clients":[]}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "with connected client",
			setupClients: true,
			expectedJSON: `{"clients":["test-client"]}`,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupClients {
				ts.wsServer.clients["test-client"] = &websocket.Conn{}
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/clients", nil)
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}
