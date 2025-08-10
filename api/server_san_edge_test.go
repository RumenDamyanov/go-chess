package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"go.rumenx.com/chess/config"
)

// Test en passant capture appears as standard pawn capture SAN (exd6)
func TestPGNSANEnPassant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	s := NewServer(cfg)
	r := gin.New()
	s.SetupRoutes(r)

	// Create game
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create game failed: %d", createRec.Code)
	}
	id := "1" // first game id

	// Move sequence: 1. e4 Nf6 2. e5 d5 3. exd6 e.p.
	// Coordinates: e2e4 g8f6 e4e5 d7d5 e5d6
	moves := []string{"e2e4", "g8f6", "e4e5", "d7d5", "e5d6"}
	for _, mv := range moves {
		body := []byte(`{"from":"` + mv[:2] + `","to":"` + mv[2:] + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("move %s failed: %d", mv, rec.Code)
		}
	}

	// Fetch PGN
	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()

	// Expect move 3. exd6 present (standard SAN for en passant capture)
	if !regexp.MustCompile(`3\.\s+exd6`).MatchString(pgn) {
		t.Errorf("expected en passant SAN 'exd6' in PGN, got: %s", pgn)
	}
}

// Test promotion SAN formatting (e8=Q with optional + if giving check)
func TestPGNSANPromotion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	s := NewServer(cfg)
	r := gin.New()
	s.SetupRoutes(r)

	// Create game
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create game failed: %d", createRec.Code)
	}
	id := "1"

	// Load custom FEN with white pawn on e7 ready to promote and black king on a8
	fen := `{"fen":"k7/4P3/8/8/8/8/8/4K3 w - - 0 1"}`
	fenReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/fen", bytes.NewBuffer([]byte(fen)))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRec := httptest.NewRecorder()
	r.ServeHTTP(fenRec, fenReq)
	if fenRec.Code != http.StatusOK {
		t.Fatalf("FEN load failed: %d", fenRec.Code)
	}

	// Promotion move e7e8Q
	promoBody := []byte(`{"from":"e7","to":"e8","promotion":"Q"}`)
	promoReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(promoBody))
	promoReq.Header.Set("Content-Type", "application/json")
	promoRec := httptest.NewRecorder()
	r.ServeHTTP(promoRec, promoReq)
	if promoRec.Code != http.StatusOK {
		t.Fatalf("promotion move failed: %d", promoRec.Code)
	}

	// Fetch PGN
	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()

	// Allow optional check indicator + or #
	if !regexp.MustCompile(`e8=Q[+#]?`).MatchString(pgn) {
		t.Errorf("expected promotion SAN 'e8=Q' (with optional check) in PGN, got: %s", pgn)
	}
}
