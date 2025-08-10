package api

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.rumenx.com/chess/config"
)

// TestWebSocketConnection verifies the websocket endpoint upgrades and returns initial game state.
func TestWebSocketConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	srv := NewServer(cfg)
	r := gin.New()
	srv.SetupRoutes(r)

	// Create a game first via HTTP POST
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/games", nil)
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("expected 201 creating game, got %d", w.Code)
	}

	// Parse created game id from JSON
	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.ID == 0 {
		t.Fatalf("expected non-zero game id")
	}

	// Start test server for websocket dialing
	ts := httptest.NewServer(r)
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	wsURL := url.URL{Scheme: "ws", Host: u.Host, Path: "/ws/games/" + strconv.Itoa(resp.ID)}

	c, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer c.Close()

	// Read initial message
	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	var initial map[string]interface{}
	if err := c.ReadJSON(&initial); err != nil {
		t.Fatalf("read initial: %v", err)
	}
	if initial["id"].(float64) != float64(resp.ID) {
		t.Fatalf("expected id %d, got %v", resp.ID, initial["id"])
	}

	// Send echo payload
	msg := map[string]interface{}{"ping": "pong"}
	if err := c.WriteJSON(msg); err != nil {
		t.Fatalf("write json: %v", err)
	}
	var echo map[string]interface{}
	if err := c.ReadJSON(&echo); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if echo["ping"] != "pong" {
		t.Fatalf("expected echo pong, got %v", echo["ping"])
	}
}
