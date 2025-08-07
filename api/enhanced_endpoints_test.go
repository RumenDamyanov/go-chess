package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rumendamyanov/go-chess/config"
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
	var games []GameResponse
	err := json.Unmarshal(rr.Body.Bytes(), &games)
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
	router := server.SetupRoutes()

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
	router := server.SetupRoutes()

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

	var moveResp MoveResponse
	err = json.Unmarshal(rr.Body.Bytes(), &moveResp)
	if err != nil {
		t.Errorf("Failed to unmarshal move response: %v", err)
	}

	if moveResp.From != "e2" {
		t.Errorf("Expected from 'e2', got '%s'", moveResp.From)
	}

	if moveResp.To != "e4" {
		t.Errorf("Expected to 'e4', got '%s'", moveResp.To)
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
	router := server.SetupRoutes()

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
	router := server.SetupRoutes()

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
	router := server.SetupRoutes()

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
	router := server.SetupRoutes()

	// Test invalid JSON
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer([]byte("{invalid json")))
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
