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

// TestPGNSANFileDisambiguation ensures minimal SAN disambiguation is applied when multiple identical pieces can reach destination
func TestPGNSANFileDisambiguation(t *testing.T) {
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

	// FEN with knights on b1 and b3 both attacking d2; they share file so we expect rank disambiguation: N1d2
	fen := `{"fen":"rnbqkbnr/pppppppp/8/8/8/1N6/PP2PPPP/RNBQK2R w KQkq - 0 1"}`
	fenReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/fen", bytes.NewBuffer([]byte(fen)))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRec := httptest.NewRecorder()
	r.ServeHTTP(fenRec, fenReq)
	if fenRec.Code != http.StatusOK {
		t.Fatalf("FEN load failed: %d", fenRec.Code)
	}

	// Move b1d2 expecting SAN N1d2
	mvBody := []byte(`{"from":"b1","to":"d2"}`)
	mvReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(mvBody))
	mvReq.Header.Set("Content-Type", "application/json")
	mvRec := httptest.NewRecorder()
	r.ServeHTTP(mvRec, mvReq)
	if mvRec.Code != http.StatusOK {
		t.Fatalf("move failed: %d", mvRec.Code)
	}

	// Fetch PGN
	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()

	if !regexp.MustCompile(`1\.\s+N1d2`).MatchString(pgn) {
		t.Errorf("expected rank-disambiguated SAN 'N1d2' in PGN, got: %s", pgn)
	}
}

// TestPGNSANRankDisambiguation ensures SAN includes rank when two identical pieces on same file different ranks
func TestPGNSANRankDisambiguation(t *testing.T) {
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

	// FEN with knights on d2 and d4 both can move to f3 -> need rank disambiguation N4f3 (since both share file d)
	fen := `{"fen":"4k3/8/8/8/3N4/8/3N4/4K3 w - - 0 1"}`
	fenReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/fen", bytes.NewBuffer([]byte(fen)))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRec := httptest.NewRecorder()
	r.ServeHTTP(fenRec, fenReq)
	if fenRec.Code != http.StatusOK {
		t.Fatalf("FEN load failed: %d", fenRec.Code)
	}

	// Move d4f3 expecting SAN N4f3
	mvBody := []byte(`{"from":"d4","to":"f3"}`)
	mvReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(mvBody))
	mvReq.Header.Set("Content-Type", "application/json")
	mvRec := httptest.NewRecorder()
	r.ServeHTTP(mvRec, mvReq)
	if mvRec.Code != http.StatusOK {
		t.Fatalf("move failed: %d", mvRec.Code)
	}

	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()

	if !regexp.MustCompile(`1\.\s+N4f3`).MatchString(pgn) {
		t.Errorf("expected rank-disambiguated SAN 'N4f3' in PGN, got: %s", pgn)
	}
}

// TestPGNSANCheckAnnotation ensures '+' added for checking move
func TestPGNSANCheckAnnotation(t *testing.T) {
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

	// FEN where Qxe7+ is possible (white queen e2 takes pawn e7 giving check to king e8)
	fen := `{"fen":"4k3/4p3/8/8/8/8/4Q3/6K1 w - - 0 1"}`
	fenReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/fen", bytes.NewBuffer([]byte(fen)))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRec := httptest.NewRecorder()
	r.ServeHTTP(fenRec, fenReq)
	if fenRec.Code != http.StatusOK {
		t.Fatalf("FEN load failed: %d", fenRec.Code)
	}

	// Move e2e7 capture
	mvBody := []byte(`{"from":"e2","to":"e7"}`)
	mvReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(mvBody))
	mvReq.Header.Set("Content-Type", "application/json")
	mvRec := httptest.NewRecorder()
	r.ServeHTTP(mvRec, mvReq)
	if mvRec.Code != http.StatusOK {
		t.Fatalf("move failed: %d", mvRec.Code)
	}

	// Fetch PGN
	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()

	if !regexp.MustCompile(`1\.\s+Qxe7\+`).MatchString(pgn) { // plus indicates check
		t.Errorf("expected checking SAN 'Qxe7+' in PGN, got: %s", pgn)
	}
}

// TestPGNSANCheckmateAnnotation ensures '#' added for a mating move
func TestPGNSANCheckmateAnnotation(t *testing.T) {
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

	// Mate pattern: Back rank style Qxf8# (queen e7 captures rook f8; white rook f1 protects queen; black king g8 trapped by own pawns and queen control)
	fen := `{"fen":"5r1k/4Q1pp/8/8/8/8/8/5RK1 w - - 0 1"}`
	fenReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/fen", bytes.NewBuffer([]byte(fen)))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRec := httptest.NewRecorder()
	r.ServeHTTP(fenRec, fenReq)
	if fenRec.Code != http.StatusOK {
		t.Fatalf("FEN load failed: %d", fenRec.Code)
	}

	mvBody := []byte(`{"from":"e7","to":"f8"}`)
	mvReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(mvBody))
	mvReq.Header.Set("Content-Type", "application/json")
	mvRec := httptest.NewRecorder()
	r.ServeHTTP(mvRec, mvReq)
	if mvRec.Code != http.StatusOK {
		t.Fatalf("move failed: %d", mvRec.Code)
	}

	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()

	if !regexp.MustCompile(`1\.\s+Qxf8#`).MatchString(pgn) {
		t.Errorf("expected mating SAN 'Qxf8#' in PGN, got: %s", pgn)
	}
}

// TestPGNSANFileDisambiguationCapture tests file disambiguation with rooks capturing same square
func TestPGNSANFileDisambiguationCapture(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	s := NewServer(cfg)
	r := gin.New()
	s.SetupRoutes(r)

	createReq := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create game failed: %d", createRec.Code)
	}
	id := "1"

	// Rooks a1 and h1 both can capture black knight on d1 -> expect Raxd1
	fen := `{"fen":"4k3/8/8/8/8/8/6K1/R2n3R w - - 0 1"}`
	fenReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/fen", bytes.NewBuffer([]byte(fen)))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRec := httptest.NewRecorder()
	r.ServeHTTP(fenRec, fenReq)
	if fenRec.Code != http.StatusOK {
		t.Fatalf("FEN load failed: %d", fenRec.Code)
	}

	mvBody := []byte(`{"from":"a1","to":"d1"}`)
	mvReq := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(mvBody))
	mvReq.Header.Set("Content-Type", "application/json")
	mvRec := httptest.NewRecorder()
	r.ServeHTTP(mvRec, mvReq)
	if mvRec.Code != http.StatusOK {
		t.Fatalf("move failed: %d", mvRec.Code)
	}

	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		t.Fatalf("PGN fetch failed: %d", pgnRec.Code)
	}
	pgn := pgnRec.Body.String()
	if !regexp.MustCompile(`1\.\s+Raxd1`).MatchString(pgn) {
		t.Errorf("expected file-disambiguated capture SAN 'Raxd1' in PGN, got: %s", pgn)
	}
}
