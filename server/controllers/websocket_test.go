package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Helper function for test setup
func setupTest() *testSetup {
	gin.SetMode(gin.TestMode)

	// Initialize logger for testing
	logger, _ := zap.NewDevelopment()

	ts := &testSetup{
		wsServer: NewWebSocketServer(WithLogger(logger)),
		router:   gin.New(),
	}
	ts.server = httptest.NewServer(ts.router)
	return ts
}

type testSetup struct {
	wsServer *WebSocketServer
	router   *gin.Engine
	server   *httptest.Server
}

// Helper function to setup a test client
func setupTestClient(ts *testSetup, clientID string) (*websocket.Conn, *http.Response, error) {
	wsURL := "ws" + strings.TrimPrefix(ts.server.URL, "http") + "/ws?client_id=" + clientID
	return websocket.DefaultDialer.Dial(wsURL, nil)
}

func TestHandleClient(t *testing.T) {
	ts := setupTest()
	ts.router.GET("/ws", ts.wsServer.HandleClient)
	defer ts.server.Close()

	tests := []struct {
		name     string
		clientID string
		wantErr  bool
	}{
		{"valid_client", "test-client", false},
		{"empty_client_id", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				resp, err := http.Get(ts.server.URL + "/ws")
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				return
			}

			// Test WebSocket connection
			ws, _, err := setupTestClient(ts, tt.clientID)
			require.NoError(t, err)
			defer ws.Close()

			// Wait for connection registration
			time.Sleep(50 * time.Millisecond)

			// Verify client connection
			ts.wsServer.mu.RLock()
			client, exists := ts.wsServer.clients[tt.clientID]
			ts.wsServer.mu.RUnlock()
			assert.True(t, exists)
			assert.NotNil(t, client)

			// Test message sending
			message := []byte("test message")
			err = ws.WriteMessage(websocket.TextMessage, message)
			require.NoError(t, err)
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
	ts.router.GET("/ws", ts.wsServer.HandleClient)
	ts.router.GET("/analytics/ws/:client_id", ts.wsServer.HandleAnalytics)
	defer ts.server.Close()

	tests := []struct {
		name     string
		clientID string
		wantErr  bool
	}{
		{"valid analytics connection", "test-client", false},
		{"invalid client id", "non-existent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				// Setup client connection first
				clientWS, _, err := setupTestClient(ts, tt.clientID)
				require.NoError(t, err)
				defer clientWS.Close()
				time.Sleep(50 * time.Millisecond)
			}

			// Test analytics connection
			httpURL := "http" + strings.TrimPrefix(ts.server.URL, "http") + "/analytics/ws/" + tt.clientID
			resp, err := http.Get(httpURL)
			require.NoError(t, err)
			defer resp.Body.Close()

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				return
			}

			// Test WebSocket connection
			wsURL := "ws" + strings.TrimPrefix(ts.server.URL, "http") + "/analytics/ws/" + tt.clientID
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			defer ws.Close()

			// Verify analytics connection
			time.Sleep(50 * time.Millisecond)
			ts.wsServer.mu.RLock()
			_, exists := ts.wsServer.analyticsConn[tt.clientID]
			ts.wsServer.mu.RUnlock()
			assert.True(t, exists)
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
