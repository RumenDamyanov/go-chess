package chat

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCreateCustomChatbot(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	tests := []struct {
		name     string
		provider string
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "OpenAI provider",
			provider: "openai",
			apiKey:   "test-openai-key",
			wantErr:  false,
		},
		{
			name:     "Anthropic provider",
			provider: "anthropic",
			apiKey:   "test-anthropic-key",
			wantErr:  false,
		},
		{
			name:     "Gemini provider",
			provider: "gemini",
			apiKey:   "test-gemini-key",
			wantErr:  false,
		},
		{
			name:     "XAI provider",
			provider: "xai",
			apiKey:   "test-xai-key",
			wantErr:  false,
		},
		{
			name:     "Invalid provider",
			provider: "invalid",
			apiKey:   "test-key",
			wantErr:  true,
		},
		{
			name:     "Empty API key",
			provider: "openai",
			apiKey:   "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatbot, err := service.createCustomChatbot(tt.provider, tt.apiKey)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for test case '%s', got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for test case '%s': %v", tt.name, err)
				return
			}

			if chatbot == nil {
				t.Errorf("Expected chatbot to be created for test case '%s'", tt.name)
			}
		})
	}
}

func TestChatWithCustomAPIKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	req := ChatRequest{
		GameID:   1,
		Message:  "What's the best opening move?",
		UserID:   "test-user",
		Provider: "openai",
		APIKey:   "test-custom-key",
		MoveData: &MoveContext{
			LastMove:   "e2e4",
			Position:   "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			LegalMoves: []string{"a7a6", "a7a5", "b7b6", "b7b5", "c7c6", "c7c5"},
			InCheck:    false,
		},
	}

	ctx := context.Background()

	// This will likely fail without a real API key, but should handle gracefully
	response, err := service.Chat(ctx, req)

	// We expect this to fail in test environment, but should not panic
	if err != nil {
		t.Logf("Expected failure without real API key: %v", err)

		// Verify the conversation was still created
		if conv, exists := service.conversations[req.GameID]; !exists {
			t.Error("Expected conversation to be created even on API failure")
		} else {
			if len(conv.Messages) == 0 {
				t.Error("Expected at least the user message to be stored")
			}
		}
	} else {
		// If it succeeds (unlikely without real key), verify response
		if response.Message == "" {
			t.Error("Expected non-empty response message")
		}
		if response.MessageID == "" {
			t.Error("Expected message ID to be set")
		}
	}
}

func TestEnhancedMoveContext(t *testing.T) {
	// Test the enhanced MoveContext structure with all new fields
	moveCtx := &MoveContext{
		LastMove:   "e2e4",
		Position:   "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		LegalMoves: []string{"a7a6", "a7a5", "b7b6", "b7b5", "c7c6", "c7c5", "d7d6", "d7d5"},
		InCheck:    false,
	}

	// Test basic fields
	if moveCtx.LastMove != "e2e4" {
		t.Errorf("Expected last move 'e2e4', got '%s'", moveCtx.LastMove)
	}

	if moveCtx.Position == "" {
		t.Error("Expected FEN position to be set")
	}

	// Test new enhanced fields
	if len(moveCtx.LegalMoves) == 0 {
		t.Error("Expected legal moves to be populated")
	}

	expectedMoves := 8
	if len(moveCtx.LegalMoves) != expectedMoves {
		t.Errorf("Expected %d legal moves, got %d", expectedMoves, len(moveCtx.LegalMoves))
	}

	if moveCtx.InCheck {
		t.Error("Expected InCheck to be false for this position")
	}
}

func TestBuildGameContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	moveData := &MoveContext{
		LastMove:      "e2e4",
		MoveCount:     1,
		CurrentPlayer: "black",
		GameStatus:    "active",
		Position:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		LegalMoves:    []string{"a7a6", "a7a5", "b7b6", "b7b5"},
		InCheck:       false,
		CapturedPiece: "",
	}

	context := service.buildGameContext(moveData)

	// Verify all expected fields are present
	if context["position_fen"] != moveData.Position {
		t.Error("Expected position_fen to be included in game context")
	}

	if context["legal_moves_count"] == nil {
		t.Error("Expected legal_moves_count to be included in game context")
	}

	if context["current_player"] != moveData.CurrentPlayer {
		t.Error("Expected current_player to be included in game context")
	}

	if context["move_count"] != moveData.MoveCount {
		t.Error("Expected move_count to be included in game context")
	}

	if context["last_move"] != moveData.LastMove {
		t.Error("Expected last_move to be included in game context")
	}

	if context["game_status"] != moveData.GameStatus {
		t.Error("Expected game_status to be included in game context")
	}
}

func TestChatResponse(t *testing.T) {
	// Test ChatResponse structure
	response := ChatResponse{
		Message:     "That's a great opening! The King's Pawn opening controls the center.",
		MessageID:   "msg-123",
		Personality: "friendly",
		GameContext: map[string]interface{}{
			"position":       "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			"legal_moves":    []string{"a7a6", "a7a5", "b7b6", "b7b5"},
			"in_check":       false,
			"current_player": "black",
		},
		Suggestions: []string{
			"Consider developing your knights first",
			"Control the center with your pawns",
			"Castle early for king safety",
		},
	}

	if response.Message == "" {
		t.Error("Expected non-empty message")
	}

	if response.MessageID == "" {
		t.Error("Expected message ID to be set")
	}

	if response.Personality == "" {
		t.Error("Expected personality to be set")
	}

	if response.GameContext == nil {
		t.Error("Expected game context to be set")
	}

	if len(response.Suggestions) == 0 {
		t.Error("Expected suggestions to be provided")
	}

	// Verify game context structure
	if position, ok := response.GameContext["position"].(string); !ok || position == "" {
		t.Error("Expected position in game context")
	}

	if legalMoves, ok := response.GameContext["legal_moves"].([]string); !ok || len(legalMoves) == 0 {
		t.Error("Expected legal moves in game context")
	}
}

func TestMessageStructure(t *testing.T) {
	now := time.Now()

	message := Message{
		ID:      "msg-456",
		Type:    "ai",
		Content: "I recommend developing your knight to f6.",
		GameState: map[string]interface{}{
			"position": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			"moves":    []string{"e2e4"},
		},
		Timestamp: now,
	}

	if message.ID == "" {
		t.Error("Expected message ID to be set")
	}

	if message.Type != "ai" {
		t.Errorf("Expected message type 'ai', got '%s'", message.Type)
	}

	if message.Content == "" {
		t.Error("Expected message content to be set")
	}

	if message.GameState == nil {
		t.Error("Expected game state to be set")
	}

	if message.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestConversationStructure(t *testing.T) {
	now := time.Now()

	conversation := Conversation{
		GameID:    123,
		Messages:  []Message{},
		Context:   make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if conversation.GameID != 123 {
		t.Errorf("Expected game ID 123, got %d", conversation.GameID)
	}

	if conversation.Messages == nil {
		t.Error("Expected messages slice to be initialized")
	}

	if conversation.Context == nil {
		t.Error("Expected context map to be initialized")
	}

	if conversation.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if conversation.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestAddMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	conversation := &Conversation{
		GameID:    1,
		Messages:  []Message{},
		Context:   make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	moveData := &MoveContext{
		LastMove:      "e2e4",
		MoveCount:     1,
		CurrentPlayer: "black",
		GameStatus:    "active",
		Position:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
	}

	messageID := service.addMessage(conversation, "user", "What's a good response to e4?", moveData)

	if messageID == "" {
		t.Error("Expected message ID to be returned")
	}

	if len(conversation.Messages) != 1 {
		t.Errorf("Expected 1 message in conversation, got %d", len(conversation.Messages))
	}

	message := conversation.Messages[0]
	if message.Type != "user" {
		t.Errorf("Expected message type 'user', got '%s'", message.Type)
	}

	if message.Content != "What's a good response to e4?" {
		t.Errorf("Expected specific content, got '%s'", message.Content)
	}

	if message.GameState == nil {
		t.Error("Expected game state to be included in message")
	}
}

func TestBuildContextualMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service, err := NewChatService(logger)
	if err != nil {
		t.Fatalf("Failed to create chat service: %v", err)
	}

	conversation := &Conversation{
		GameID:   1,
		Messages: []Message{},
		Context:  make(map[string]interface{}),
	}

	moveData := &MoveContext{
		LastMove:      "e2e4",
		MoveCount:     1,
		CurrentPlayer: "black",
		GameStatus:    "active",
		Position:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		LegalMoves:    []string{"e7e6", "e7e5", "d7d6", "d7d5"},
		InCheck:       false,
	}

	userMessage := "What should I play in response to e4?"
	contextualMessage := service.buildContextualMessage(userMessage, conversation, moveData)

	if contextualMessage == "" {
		t.Error("Expected non-empty contextual message")
	}

	// Should include the user message
	if !containsString(contextualMessage, userMessage) {
		t.Error("Expected contextual message to include user message")
	}

	// Should include game state information
	if !containsString(contextualMessage, "e2e4") {
		t.Error("Expected contextual message to include last move")
	}

	// Should include position information
	if !containsString(contextualMessage, "Position:") {
		t.Error("Expected contextual message to include position information")
	}
}

// Helper function for string containment checks
func containsString(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		haystack != needle &&
		findInString(haystack, needle)
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
