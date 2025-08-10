package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.rumenx.com/chess/config"
)

func TestPGNEndpointBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	s := NewServer(cfg)
	r := gin.New()
	s.SetupRoutes(r)

	// Create a new game
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	createRec := httptest.NewRecorder()
	r.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		b, _ := ioutil.ReadAll(createRec.Body)
		t.Fatalf("expected 201 create got %d body=%s", createRec.Code, string(b))
	}

	// Extract game id from response body (simple regex on id field)
	body := createRec.Body.String()
	idRe := regexp.MustCompile(`"id":\s*(\\d+)`)
	m := idRe.FindStringSubmatch(body)
	if len(m) < 2 {
		// fallback: just request PGN for id 1 (first game)
		m = []string{"", "1"}
	}
	id := m[1]

	// Make a couple of moves to populate PGN
	moves := []string{"e2e4", "e7e5", "g1f3"}
	for _, mv := range moves {
		jsonBody := []byte(`{"from":"` + mv[:2] + `","to":"` + mv[2:] + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/games/"+id+"/moves", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b, _ := ioutil.ReadAll(rec.Body)
			t.Fatalf("move %s expected 200 got %d body=%s", mv, rec.Code, string(b))
		}
	}

	// Get PGN
	pgnReq := httptest.NewRequest(http.MethodGet, "/api/games/"+id+"/pgn", nil)
	pgnRec := httptest.NewRecorder()
	r.ServeHTTP(pgnRec, pgnReq)
	if pgnRec.Code != http.StatusOK {
		b, _ := ioutil.ReadAll(pgnRec.Body)
		t.Fatalf("expected 200 PGN got %d body=%s", pgnRec.Code, string(b))
	}
	pgn := pgnRec.Body.String()

	// Basic assertions
	requiredTags := []string{"[Event ", "[Site ", "[Date ", "[White ", "[Black ", "[Result "}
	for _, tag := range requiredTags {
		if !strings.Contains(pgn, tag) {
			t.Errorf("PGN missing tag prefix %s", tag)
		}
	}
	// Accept either coordinate or SAN notation for first move (now SAN: e4)
	if !strings.Contains(pgn, "1. e4") && !strings.Contains(pgn, "1. e2e4") {
		t.Errorf("PGN missing first move sequence, got: %s", pgn)
	}
	if !strings.HasSuffix(strings.TrimSpace(pgn), "*") { // game still in progress
		// Accept alternative if result already set differently
		if !(strings.HasSuffix(strings.TrimSpace(pgn), "1-0") || strings.HasSuffix(strings.TrimSpace(pgn), "0-1") || strings.HasSuffix(strings.TrimSpace(pgn), "1/2-1/2")) {
			t.Errorf("PGN termination marker missing or invalid: %q", pgn)
		}
	}
}
