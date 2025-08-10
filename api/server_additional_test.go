package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.rumenx.com/chess/config"
)

// Helper to set up test server and router.
func newTestServerAndRouter() (*Server, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	s := NewServer(cfg)
	r := gin.New()
	s.SetupRoutes(r)
	return s, r
}

func createGame(t *testing.T, r *gin.Engine) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create game status %d", rec.Code)
	}
	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return resp.ID
}

// Minimal itoa to avoid strconv import noise (json.Number.String uses underlying string) – we can just use Sprintf but keep small.
func itoa(i int) string { return strconv.Itoa(i) }

// Test getAIMove when it's not the AI's turn (should return not_ai_turn error).
func TestGetAIMove_NotAITurn(t *testing.T) {
	_, r := newTestServerAndRouter()
	id := createGame(t, r)
	body := []byte(`{"level":"medium","engine":"random"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+itoa(id)+"/ai-move", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 not_ai_turn, got %d", rec.Code)
	}
}

// Test loadFromFEN invalid input branch.
func TestLoadFromFEN_Invalid(t *testing.T) {
	_, r := newTestServerAndRouter()
	id := createGame(t, r)
	body := []byte(`{"fen":"invalid fen"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+itoa(id)+"/fen", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid_fen, got %d", rec.Code)
	}
}

// Test analyzePosition basic response shape.
func TestAnalyzePosition_Basic(t *testing.T) {
	_, r := newTestServerAndRouter()
	id := createGame(t, r)
	req := httptest.NewRequest(http.MethodGet, "/api/games/"+itoa(id)+"/analysis", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 analysis, got %d", rec.Code)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, k := range []string{"status", "active_color", "move_count", "evaluation"} {
		if _, ok := data[k]; !ok {
			t.Fatalf("missing key %s in analysis", k)
		}
	}
}

// Test getAIHint fallback error path by requesting a hint; accept success or deterministic fallback.
func TestGetAIHint_FallbackOrSuccess(t *testing.T) {
	_, r := newTestServerAndRouter()
	id := createGame(t, r)
	body := []byte(`{"level":"medium","engine":"minimax"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+itoa(id)+"/ai-hint", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	start := time.Now()
	r.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		return
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("deterministic")) {
		t.Fatalf("expected deterministic flag in fallback body: %s", rec.Body.String())
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("hint request took unexpectedly long")
	}
}
