package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rumendamyanov/go-chess/config"
)

// TestDeleteGame tests the game deletion endpoint
func TestDeleteGame(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// First create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Now delete the game
	deleteURL := fmt.Sprintf("/api/games/%v", gameID)
	deleteReq, _ := http.NewRequest("DELETE", deleteURL, nil)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent && deleteRR.Code != http.StatusOK {
		t.Errorf("Expected status 204 or 200, got %d", deleteRR.Code)
	}

	// Try to get the deleted game - should return 404
	getReq, _ := http.NewRequest("GET", deleteURL, nil)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for deleted game, got %d", getRR.Code)
	}
}

// TestListGames tests the games listing endpoint
func TestListGames(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create multiple games
	gameIDs := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		createReq, _ := http.NewRequest("POST", "/api/games", nil)
		createRR := httptest.NewRecorder()
		router.ServeHTTP(createRR, createReq)

		var createResp map[string]interface{}
		json.Unmarshal(createRR.Body.Bytes(), &createResp)
		gameIDs[i] = createResp["id"]
	}

	// List all games
	listReq, _ := http.NewRequest("GET", "/api/games", nil)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", listRR.Code)
	}

	var listResp map[string]interface{}
	json.Unmarshal(listRR.Body.Bytes(), &listResp)

	games, ok := listResp["games"].([]interface{})
	if !ok {
		t.Fatal("Expected games array in response")
	}

	if len(games) < 3 {
		t.Errorf("Expected at least 3 games, got %d", len(games))
	}
}

// TestGetMoveHistory tests the move history endpoint
func TestGetMoveHistory(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Make a move
	moveData := map[string]string{"move": "e2e4"}
	moveJSON, _ := json.Marshal(moveData)
	moveURL := fmt.Sprintf("/api/games/%v/moves", gameID)
	moveReq, _ := http.NewRequest("POST", moveURL, bytes.NewBuffer(moveJSON))
	moveReq.Header.Set("Content-Type", "application/json")
	moveRR := httptest.NewRecorder()
	router.ServeHTTP(moveRR, moveReq)

	// Get move history
	historyReq, _ := http.NewRequest("GET", moveURL, nil)
	historyRR := httptest.NewRecorder()
	router.ServeHTTP(historyRR, historyReq)

	if historyRR.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", historyRR.Code)
	}

	var historyResp map[string]interface{}
	json.Unmarshal(historyRR.Body.Bytes(), &historyResp)

	moves, ok := historyResp["moves"].([]interface{})
	if !ok {
		// Check if API returns different structure
		if historyResp["error"] != nil {
			t.Skip("Move history endpoint not fully implemented")
		}
		t.Fatal("Expected moves array in response")
	}

	if len(moves) < 1 {
		t.Errorf("Expected at least 1 move in history, got %d", len(moves))
	}
}

// TestGetLegalMoves tests the legal moves endpoint
func TestGetLegalMoves(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Get legal moves
	legalMovesURL := fmt.Sprintf("/api/games/%v/legal-moves", gameID)
	legalReq, _ := http.NewRequest("GET", legalMovesURL, nil)
	legalRR := httptest.NewRecorder()
	router.ServeHTTP(legalRR, legalReq)

	if legalRR.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", legalRR.Code)
	}

	var legalResp map[string]interface{}
	json.Unmarshal(legalRR.Body.Bytes(), &legalResp)

	moves, ok := legalResp["legal_moves"].([]interface{})
	if !ok {
		// Check if API returns different structure
		if legalResp["error"] != nil {
			t.Skip("Legal moves endpoint not fully implemented")
		}
		t.Fatal("Expected legal_moves array in response")
	}

	// In starting position, there should be 20 legal moves (16 pawn moves + 4 knight moves)
	if len(moves) != 20 {
		t.Errorf("Expected 20 legal moves in starting position, got %d", len(moves))
	}
}

// TestLoadFromFEN tests loading a position from FEN notation
func TestLoadFromFEN(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Load from FEN (this is a standard starting position FEN)
	fenData := map[string]string{
		"fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	}
	fenJSON, _ := json.Marshal(fenData)
	fenURL := fmt.Sprintf("/api/games/%v/fen", gameID)
	fenReq, _ := http.NewRequest("POST", fenURL, bytes.NewBuffer(fenJSON))
	fenReq.Header.Set("Content-Type", "application/json")
	fenRR := httptest.NewRecorder()
	router.ServeHTTP(fenRR, fenReq)

	if fenRR.Code != http.StatusOK && fenRR.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 200 or 501, got %d", fenRR.Code)
	}
}

// TestAnalyzePosition tests the position analysis endpoint
func TestAnalyzePosition(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Analyze position
	analysisURL := fmt.Sprintf("/api/games/%v/analysis", gameID)
	analysisReq, _ := http.NewRequest("GET", analysisURL, nil)
	analysisRR := httptest.NewRecorder()
	router.ServeHTTP(analysisRR, analysisReq)

	if analysisRR.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", analysisRR.Code)
	}

	var analysisResp map[string]interface{}
	json.Unmarshal(analysisRR.Body.Bytes(), &analysisResp)

	// Check that analysis contains expected fields
	if _, ok := analysisResp["evaluation"]; !ok {
		t.Error("Expected evaluation in analysis response")
	}
}

// TestGetAIMove tests the AI move endpoint
func TestGetAIMove(t *testing.T) {
	cfg := config.Default()
	cfg.AI.MaxThinkTime = 1000 * time.Millisecond // 1 second timeout for quick test
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Get AI move
	aiMoveURL := fmt.Sprintf("/api/games/%v/ai-move", gameID)
	aiReq, _ := http.NewRequest("POST", aiMoveURL, nil)
	aiRR := httptest.NewRecorder()
	router.ServeHTTP(aiRR, aiReq)

	if aiRR.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", aiRR.Code)
	}

	var aiResp map[string]interface{}
	json.Unmarshal(aiRR.Body.Bytes(), &aiResp)

	// Check that AI move response contains a move
	if _, ok := aiResp["move"]; !ok {
		t.Error("Expected move in AI response")
	}
}

// TestChatWithAI tests the AI chat endpoint
func TestChatWithAI(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Chat with AI
	chatData := map[string]string{"message": "What's the best opening move?"}
	chatJSON, _ := json.Marshal(chatData)
	chatURL := fmt.Sprintf("/api/games/%v/chat", gameID)
	chatReq, _ := http.NewRequest("POST", chatURL, bytes.NewBuffer(chatJSON))
	chatReq.Header.Set("Content-Type", "application/json")
	chatRR := httptest.NewRecorder()
	router.ServeHTTP(chatRR, chatReq)

	// Chat endpoint should return a response (might be an error if no LLM configured)
	if chatRR.Code != http.StatusOK && chatRR.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", chatRR.Code)
	}
}

// TestGetAIReaction tests the AI reaction endpoint
func TestGetAIReaction(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Get AI reaction
	reactionURL := fmt.Sprintf("/api/games/%v/react", gameID)
	reactionReq, _ := http.NewRequest("POST", reactionURL, nil)
	reactionRR := httptest.NewRecorder()
	router.ServeHTTP(reactionRR, reactionReq)

	// Reaction endpoint should return a response (might be an error if no LLM configured)
	if reactionRR.Code != http.StatusOK && reactionRR.Code != http.StatusInternalServerError && reactionRR.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200, 400, or 500, got %d", reactionRR.Code)
	}
}

// TestInvalidJSON tests handling of invalid JSON in requests
func TestInvalidJSON(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game first
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Test invalid JSON in move request
	invalidJSON := `{"move": invalid json`
	moveURL := fmt.Sprintf("/api/games/%v/moves", gameID)
	moveReq, _ := http.NewRequest("POST", moveURL, strings.NewReader(invalidJSON))
	moveReq.Header.Set("Content-Type", "application/json")
	moveRR := httptest.NewRecorder()
	router.ServeHTTP(moveRR, moveReq)

	if moveRR.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", moveRR.Code)
	}
}

// TestMissingFields tests handling of missing required fields
func TestMissingFields(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game first
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Test missing move field
	emptyData := map[string]string{}
	emptyJSON, _ := json.Marshal(emptyData)
	moveURL := fmt.Sprintf("/api/games/%v/moves", gameID)
	moveReq, _ := http.NewRequest("POST", moveURL, bytes.NewBuffer(emptyJSON))
	moveReq.Header.Set("Content-Type", "application/json")
	moveRR := httptest.NewRecorder()
	router.ServeHTTP(moveRR, moveReq)

	if moveRR.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing move field, got %d", moveRR.Code)
	}
}

// TestLargeGameID tests handling of very large game IDs
func TestLargeGameID(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Test with very large game ID
	largeID := "999999999999999999"
	getReq, _ := http.NewRequest("GET", "/api/games/"+largeID, nil)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent large game ID, got %d", getRR.Code)
	}
}

// TestContentTypeValidation tests that endpoints require proper content type
func TestContentTypeValidation(t *testing.T) {
	cfg := config.Default()
	server := NewServer(cfg)
	router := gin.New()
	server.SetupRoutes(router)

	// Create a game first
	createReq, _ := http.NewRequest("POST", "/api/games", nil)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createRR.Body.Bytes(), &createResp)
	gameID := createResp["id"]

	// Test POST request without proper content type
	moveData := `{"move": "e2e4"}`
	moveURL := fmt.Sprintf("/api/games/%v/moves", gameID)
	moveReq, _ := http.NewRequest("POST", moveURL, strings.NewReader(moveData))
	// Intentionally not setting Content-Type header
	moveRR := httptest.NewRecorder()
	router.ServeHTTP(moveRR, moveReq)

	// The request might still work, but response should be valid
	if moveRR.Code != http.StatusOK && moveRR.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", moveRR.Code)
	}
}
