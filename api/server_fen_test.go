package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"go.rumenx.com/chess/config"
)

// helper to setup router
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	s := NewServer(cfg)
	r := gin.New()
	s.SetupRoutes(r)
	return r
}

func TestLoadFromFENEndpoint(t *testing.T) {
	r := setupTestRouter()

	// Create a game
	req := httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBufferString(`{"ai_color":"black"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 create, got %d", w.Code)
	}
	var createResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	id := int(createResp["id"].(float64))

	// Load FEN
	fenBody := `{"fen":"8/8/8/8/8/8/8/8 w - - 0 1"}`
	loadReq := httptest.NewRequest(http.MethodPost, "/api/games/"+strconv.Itoa(id)+"/fen", bytes.NewBufferString(fenBody))
	loadReq.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, loadReq)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 from FEN load, got %d body=%s", w2.Code, w2.Body.String())
	}
	var loadResp map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &loadResp); err != nil {
		t.Fatalf("unmarshal load response: %v", err)
	}
	if loadResp["move_count"].(float64) != 1 {
		t.Fatalf("expected move_count 1 after FEN load, got %v", loadResp["move_count"])
	}
	if loadResp["active_color"].(string) != "white" {
		t.Fatalf("expected active_color white, got %s", loadResp["active_color"])
	}

	// Attempt invalid FEN
	badBody := `{"fen":"invalid"}`
	badReq := httptest.NewRequest(http.MethodPost, "/api/games/"+strconv.Itoa(id)+"/fen", bytes.NewBufferString(badBody))
	badReq.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, badReq)
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad FEN, got %d", w3.Code)
	}
}
