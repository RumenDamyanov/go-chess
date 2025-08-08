package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.rumenx.com/chess/config"
)

func TestServerInitialization(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	if server == nil {
		t.Error("Expected server to be created")
	}

	if server.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestGameRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	router := gin.New()
	server.SetupRoutes(router)
	if router == nil {
		t.Error("Expected router to be created")
	}

	// Test games endpoint
	req := httptest.NewRequest("GET", "/api/games", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Should return JSON
	var gamesResponse struct {
		Games []GameResponse `json:"games"`
		Count int            `json:"count"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &gamesResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal games response: %v", err)
	}
}

func TestGameCreation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server.SetupRoutes(router)

	// Test create game
	createReq := GameCreateRequest{
		AIColor: "black",
	}

	jsonBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal create request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var game GameResponse
	err = json.Unmarshal(rr.Body.Bytes(), &game)
	if err != nil {
		t.Errorf("Failed to unmarshal game response: %v", err)
	}

	if game.ID == 0 {
		t.Error("Expected game to have an ID")
	}

	if game.Status == "" {
		t.Error("Expected game to have a status")
	}
}

func TestMoveValidation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server.SetupRoutes(router)

	// First create a game
	createReq := GameCreateRequest{AIColor: "black"}
	jsonBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var game GameResponse
	json.Unmarshal(rr.Body.Bytes(), &game)

	// Test valid move
	moveReq := MoveRequest{
		From: "e2",
		To:   "e4",
	}

	jsonBody, err := json.Marshal(moveReq)
	if err != nil {
		t.Fatalf("Failed to marshal move request: %v", err)
	}

	req = httptest.NewRequest("POST", "/api/games/1/moves", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid move, got %d: %s", rr.Code, rr.Body.String())
	}

	var gameResp GameResponse
	err = json.Unmarshal(rr.Body.Bytes(), &gameResp)
	if err != nil {
		t.Errorf("Failed to unmarshal game response: %v", err)
	}

	// Check the last move in the move history
	if len(gameResp.MoveHistory) == 0 {
		t.Errorf("Expected move history to contain at least one move")
		return
	}

	lastMove := gameResp.MoveHistory[len(gameResp.MoveHistory)-1]
	if lastMove.From != "e2" {
		t.Errorf("Expected from 'e2', got '%s'", lastMove.From)
	}

	if lastMove.To != "e4" {
		t.Errorf("Expected to 'e4', got '%s'", lastMove.To)
	}
}

func TestInvalidMoveHandling(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game first
	createReq := GameCreateRequest{AIColor: "black"}
	jsonBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Test invalid move
	moveReq := MoveRequest{
		From: "e2",
		To:   "e5", // Invalid - can't move pawn two squares to e5
	}

	jsonBody, err := json.Marshal(moveReq)
	if err != nil {
		t.Fatalf("Failed to marshal move request: %v", err)
	}

	req = httptest.NewRequest("POST", "/api/games/1/moves", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid move, got %d", rr.Code)
	}

	var errorResp ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &errorResp)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Error == "" {
		t.Error("Expected error message for invalid move")
	}
}

func TestGameStateRetrieval(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game first
	createReq := GameCreateRequest{AIColor: "black"}
	jsonBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Get game state
	req = httptest.NewRequest("GET", "/api/games/1", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var game GameResponse
	err := json.Unmarshal(rr.Body.Bytes(), &game)
	if err != nil {
		t.Errorf("Failed to unmarshal game response: %v", err)
	}

	if game.ID != 1 {
		t.Errorf("Expected game ID 1, got %d", game.ID)
	}

	if game.Board == "" {
		t.Error("Expected board representation in response")
	}

	if game.Status == "" {
		t.Error("Expected game status in response")
	}

	if game.MoveHistory == nil {
		t.Error("Expected move history to be initialized")
	}
}

func TestChatIntegration(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game first
	createReq := GameCreateRequest{AIColor: "black"}
	jsonBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Test chat endpoint
	chatReq := ChatRequest{
		Message:  "What's the best opening move?",
		Provider: "openai",
		APIKey:   "test-key",
	}

	jsonBody, err := json.Marshal(chatReq)
	if err != nil {
		t.Fatalf("Failed to marshal chat request: %v", err)
	}

	req = httptest.NewRequest("POST", "/api/games/1/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should not crash - might return 500 due to no real API key
	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected 200 or 500, got %d: %s", rr.Code, rr.Body.String())
	}

	// Response should be JSON even on error
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Logf("Expected JSON content type, got %s", contentType)
	}
}

func TestJSONHandling(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	server := NewServer(cfg)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server.SetupRoutes(router)

	// Test invalid JSON on move endpoint (which requires strict JSON)
	// First create a game
	createReq := httptest.NewRequest("POST", "/api/games", nil)
	createRr := httptest.NewRecorder()
	router.ServeHTTP(createRr, createReq)

	if createRr.Code != http.StatusCreated {
		t.Fatalf("Failed to create game for JSON test: %d", createRr.Code)
	}

	// Now test invalid JSON on move endpoint
	req := httptest.NewRequest("POST", "/api/games/1/moves", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", rr.Code)
	}

	var errorResp ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Error == "" {
		t.Error("Expected error message for invalid JSON")
	}
}
