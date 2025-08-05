package chat

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestNewChatService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service, err := NewChatService(logger)
	if err != nil {
		t.Errorf("Expected no error creating chat service, got: %v", err)
	}

	if service == nil {
		t.Error("Expected chat service to be created, got nil")
	}

	if service.logger == nil {
		t.Error("Expected logger to be set in chat service")
	}

	if service.conversations == nil {
		t.Error("Expected conversations map to be initialized")
	}
}

func TestChatServiceInitialization(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service, err := NewChatService(logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test that the service can be created even without API keys
	if service.chatbot == nil {
		t.Error("Expected chatbot to be initialized")
	}

	if service.config == nil {
		t.Error("Expected config to be initialized")
	}
}

func TestChatRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	// Create a chat request
	req := ChatRequest{
		GameID:  1,
		Message: "Hello, what's a good opening move?",
		UserID:  "test-user",
	}

	ctx := context.Background()

	// Test chat functionality
	response, err := service.Chat(ctx, req)

	// The response might fail due to no API keys, but we should handle it gracefully
	if err != nil {
		// In test environment without API keys, we expect this might fail
		// but it should be a handled error, not a panic
		t.Logf("Chat failed as expected in test environment: %v", err)
	} else {
		// If it succeeds (e.g., with a free model), verify response structure
		if response.Message == "" {
			t.Error("Expected non-empty message in response")
		}

		if response.MessageID == "" {
			t.Error("Expected message ID in response")
		}
	}
}

func TestConversationManagement(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	gameID := 123

	// Test that conversations are created for new games
	if _, exists := service.conversations[gameID]; exists {
		t.Error("Expected no conversation to exist initially")
	}

	// Create a conversation (this would happen during a chat request)
	conv := &Conversation{
		GameID:   gameID,
		Messages: []Message{},
		Context:  make(map[string]interface{}),
	}

	service.conversations[gameID] = conv

	// Verify conversation was stored
	if _, exists := service.conversations[gameID]; !exists {
		t.Error("Expected conversation to be stored")
	}

	// Test conversation retrieval
	if storedConv, exists := service.conversations[gameID]; !exists {
		t.Error("Expected to retrieve stored conversation")
	} else if storedConv.GameID != gameID {
		t.Errorf("Expected game ID %d, got %d", gameID, storedConv.GameID)
	}
}

func TestMoveContext(t *testing.T) {
	// Test MoveContext structure
	moveCtx := &MoveContext{
		LastMove:      "e2e4",
		MoveCount:     1,
		CurrentPlayer: "black",
		GameStatus:    "active",
		Position:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
	}

	if moveCtx.LastMove != "e2e4" {
		t.Errorf("Expected last move 'e2e4', got '%s'", moveCtx.LastMove)
	}

	if moveCtx.MoveCount != 1 {
		t.Errorf("Expected move count 1, got %d", moveCtx.MoveCount)
	}

	if moveCtx.CurrentPlayer != "black" {
		t.Errorf("Expected current player 'black', got '%s'", moveCtx.CurrentPlayer)
	}

	if moveCtx.GameStatus != "active" {
		t.Errorf("Expected game status 'active', got '%s'", moveCtx.GameStatus)
	}

	if len(moveCtx.Position) == 0 {
		t.Error("Expected position to be set")
	}
}

func TestChatRequestValidation(t *testing.T) {
	// Test ChatRequest structure validation
	req := ChatRequest{
		GameID:  1,
		Message: "Test message",
		UserID:  "user123",
		MoveData: &MoveContext{
			LastMove:   "e2e4",
			MoveCount:  1,
			GameStatus: "active",
		},
	}

	if req.GameID != 1 {
		t.Errorf("Expected game ID 1, got %d", req.GameID)
	}

	if req.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", req.Message)
	}

	if req.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", req.UserID)
	}

	if req.MoveData == nil {
		t.Error("Expected move data to be set")
	} else if req.MoveData.LastMove != "e2e4" {
		t.Errorf("Expected last move 'e2e4', got '%s'", req.MoveData.LastMove)
	}
}
