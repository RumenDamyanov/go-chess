package ai

import (
	"strings"
	"testing"

	"github.com/rumendamyanov/go-chess/engine"
)

func TestLLMAIEngineCreation(t *testing.T) {
	// Test LLMAIEngine creation with OpenAI
	config := LLMConfig{
		Provider:    ProviderOpenAI,
		APIKey:      "test-key",
		Model:       "gpt-4",
		Difficulty:  DifficultyMedium,
		Personality: "friendly",
		ChatEnabled: true,
	}

	llmEngine, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create LLM engine: %v", err)
	}

	if llmEngine == nil {
		t.Error("Expected LLM engine to be created")
	}

	if llmEngine.config.Provider != ProviderOpenAI {
		t.Errorf("Expected provider OpenAI, got %v", llmEngine.config.Provider)
	}
}

func TestLLMAIFromEnv(t *testing.T) {
	// Test creating LLM AI from environment
	llmEngine, err := NewLLMAIFromEnv("openai", DifficultyMedium)

	// This might fail without env vars, but should handle gracefully
	if err != nil {
		t.Logf("Expected failure without env vars: %v", err)

		// Error should be descriptive
		if err.Error() == "" {
			t.Error("Expected descriptive error message")
		}
	} else {
		// If it succeeds, verify basic properties
		if llmEngine == nil {
			t.Error("Expected LLM engine to be created")
		}
	}
}

func TestMakeMove(t *testing.T) {
	game := engine.NewGame()

	// Test valid move using ParseMove
	move, err := game.ParseMove("e2e4")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	err = game.MakeMove(move)
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

	// Verify game state changed
	if game.MoveCount() != 1 { // Should be move 1 after white's first move
		t.Errorf("Expected move count 1, got %d", game.MoveCount())
	}

	if game.ActiveColor() != engine.Black {
		t.Error("Expected black to be active after white's move")
	}
}

func TestGameStateAccess(t *testing.T) {
	game := engine.NewGame()

	// Test accessing board
	board := game.Board()
	if board == nil {
		t.Error("Expected game board to be accessible")
	}

	// Test FEN notation
	fen := game.ToFEN()
	if fen == "" {
		t.Error("Expected FEN notation to be available")
	}

	// Verify initial FEN
	expectedInitialFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if fen != expectedInitialFEN {
		t.Errorf("Expected initial FEN '%s', got '%s'", expectedInitialFEN, fen)
	}
}

func TestMultipleMoveSequence(t *testing.T) {
	game := engine.NewGame()

	// Test a sequence of moves
	moveSequence := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}

	for i, moveStr := range moveSequence {
		move, err := game.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %d (%s): %v", i+1, moveStr, err)
		}

		err = game.MakeMove(move)
		if err != nil {
			t.Fatalf("Failed to make move %d (%s): %v", i+1, moveStr, err)
		}

		// Verify game state is consistent
		fen := game.ToFEN()
		if fen == "" {
			t.Errorf("Expected FEN after move %d", i+1)
		}

		// Verify whose turn it is
		expectedColor := engine.Black
		if i%2 == 0 { // After white's move (even index), it's black's turn
			expectedColor = engine.Black
		} else { // After black's move (odd index), it's white's turn
			expectedColor = engine.White
		}

		actualColor := game.ActiveColor()
		if actualColor != expectedColor {
			t.Errorf("After move %d, expected %v to move, got %v", i+1, expectedColor, actualColor)
		}
	}

	// Verify final position
	finalFEN := game.ToFEN()
	if !strings.ContainsAny(finalFEN, "rnbqk") { // Should have black pieces
		t.Errorf("Final FEN should contain black pieces notation, got: %s", finalFEN)
	}

	if !strings.ContainsAny(finalFEN, "RNBQK") { // Should have white pieces
		t.Errorf("Final FEN should contain white pieces notation, got: %s", finalFEN)
	}
}

func TestLLMConfigValidation(t *testing.T) {
	// Test valid configs for different providers
	providers := []LLMProvider{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGemini,
		ProviderXAI,
		ProviderDeepSeek,
	}

	for _, provider := range providers {
		config := LLMConfig{
			Provider:    provider,
			APIKey:      "test-key",
			Model:       "test-model",
			Difficulty:  DifficultyEasy,
			Personality: "helpful",
			ChatEnabled: true,
		}

		llmEngine, err := NewLLMAIEngine(config)
		if err != nil {
			t.Errorf("Failed to create LLM engine for provider %v: %v", provider, err)
			continue
		}

		if llmEngine == nil {
			t.Errorf("Expected LLM engine to be created for provider %v", provider)
		}
	}
}

func TestDifficultyLevels(t *testing.T) {
	difficulties := []Difficulty{DifficultyEasy, DifficultyMedium, DifficultyHard, DifficultyExpert}

	for _, difficulty := range difficulties {
		config := LLMConfig{
			Provider:    ProviderOpenAI,
			APIKey:      "test-key",
			Model:       "gpt-4",
			Difficulty:  difficulty,
			Personality: "analytical",
			ChatEnabled: true,
		}

		llmEngine, err := NewLLMAIEngine(config)
		if err != nil {
			t.Errorf("Failed to create LLM engine for difficulty %v: %v", difficulty, err)
			continue
		}

		if llmEngine.config.Difficulty != difficulty {
			t.Errorf("Expected difficulty %v, got %v", difficulty, llmEngine.config.Difficulty)
		}
	}
}

func TestChatMessageStructure(t *testing.T) {
	message := ChatMessage{
		Role:    "user",
		Content: "What's the best opening move in chess?",
	}

	if message.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", message.Role)
	}

	if message.Content == "" {
		t.Error("Expected message content to be set")
	}

	// Test assistant message
	assistantMessage := ChatMessage{
		Role:    "assistant",
		Content: "1.e4 is considered one of the best opening moves.",
	}

	if assistantMessage.Role != "assistant" {
		t.Error("Expected assistant role")
	}
}
