package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rumendamyanov/go-chess/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	if server == nil {
		t.Error("NewServer returned nil")
	}
}

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Just check that it contains "status" field - the actual response is more detailed
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if _, ok := response["status"]; !ok {
		t.Error("Health endpoint should contain 'status' field")
	}
}

func TestCreateGame(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	req, err := http.NewRequest("POST", "/api/games", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Create game returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if _, ok := response["id"]; !ok {
		t.Error("Response should contain 'id' field")
	}

	if _, ok := response["board"]; !ok {
		t.Error("Response should contain 'board' field")
	}
}

func TestGetGame(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	// First create a game
	req, err := http.NewRequest("POST", "/api/games", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var createResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &createResponse); err != nil {
		t.Fatal(err)
	}

	gameID := fmt.Sprintf("%.0f", createResponse["id"].(float64))

	// Now get the game
	req, err = http.NewRequest("GET", "/api/games/"+gameID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Get game returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if fmt.Sprintf("%.0f", response["id"].(float64)) != gameID {
		t.Errorf("Response ID mismatch: got %v want %v", response["id"], gameID)
	}
}

func TestMakeMove(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	// First create a game
	req, err := http.NewRequest("POST", "/api/games", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var createResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &createResponse); err != nil {
		t.Fatal(err)
	}

	gameID := fmt.Sprintf("%.0f", createResponse["id"].(float64))

	// Make a move
	moveData := map[string]string{
		"from": "e2",
		"to":   "e4",
	}
	moveJSON, _ := json.Marshal(moveData)

	req, err = http.NewRequest("POST", "/api/games/"+gameID+"/moves", bytes.NewBuffer(moveJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Make move returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if _, ok := response["board"]; !ok {
		t.Error("Response should contain 'board' field")
	}
}

func TestCORSHeaders(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	req, err := http.NewRequest("OPTIONS", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check CORS headers
	if corsOrigin := rr.Header().Get("Access-Control-Allow-Origin"); corsOrigin == "" {
		t.Error("CORS Access-Control-Allow-Origin header not set")
	}

	if corsHeaders := rr.Header().Get("Access-Control-Allow-Headers"); corsHeaders == "" {
		t.Error("CORS Access-Control-Allow-Headers header not set")
	}
}

func TestInvalidGameID(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	// Try to get a non-existent game
	req, err := http.NewRequest("GET", "/api/games/invalid-id", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Invalid game ID should return 400, got %v", status)
	}
}

func TestJSONErrorResponse(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	// Send invalid JSON to move endpoint
	invalidJSON := `{"from": "e2", "to":}`
	req, err := http.NewRequest("POST", "/api/games/test-id/moves", bytes.NewBufferString(invalidJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Invalid JSON should return 400, got %v", status)
	}
}

func TestMultipleGames(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)

	router := gin.New()
	server.SetupRoutes(router)

	// Create multiple games
	gameIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("POST", "/api/games", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}

		gameIDs[i] = fmt.Sprintf("%.0f", response["id"].(float64))
	}

	// Verify all games exist and have unique IDs
	uniqueIDs := make(map[string]bool)
	for _, id := range gameIDs {
		if uniqueIDs[id] {
			t.Errorf("Duplicate game ID: %s", id)
		}
		uniqueIDs[id] = true

		// Verify game can be retrieved
		req, err := http.NewRequest("GET", "/api/games/"+id, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Game %s not found", id)
		}
	}
}
