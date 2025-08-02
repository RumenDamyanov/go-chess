package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rumendamyanov/go-chess/engine"
)

func TestNewLLMAIEngine(t *testing.T) {
	tests := []struct {
		name    string
		config  LLMConfig
		wantErr bool
	}{
		{
			name: "valid OpenAI config",
			config: LLMConfig{
				Provider:    ProviderOpenAI,
				APIKey:      "test-key",
				Model:       "gpt-3.5-turbo",
				Difficulty:  DifficultyMedium,
				ChatEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: LLMConfig{
				Provider:   ProviderOpenAI,
				Difficulty: DifficultyMedium,
			},
			wantErr: true,
		},
		{
			name: "valid Anthropic config",
			config: LLMConfig{
				Provider:    ProviderAnthropic,
				APIKey:      "test-key",
				Difficulty:  DifficultyHard,
				ChatEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "valid Gemini config",
			config: LLMConfig{
				Provider:    ProviderGemini,
				APIKey:      "test-key",
				Difficulty:  DifficultyEasy,
				ChatEnabled: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai, err := NewLLMAIEngine(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLLMAIEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ai == nil {
				t.Error("NewLLMAIEngine() returned nil AI engine")
			}
		})
	}
}

func TestLLMAIEngine_GetDifficulty(t *testing.T) {
	config := LLMConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		Difficulty: DifficultyHard,
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	if ai.GetDifficulty() != DifficultyHard {
		t.Errorf("Expected difficulty %v, got %v", DifficultyHard, ai.GetDifficulty())
	}
}

func TestLLMAIEngine_SetDifficulty(t *testing.T) {
	config := LLMConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		Difficulty: DifficultyEasy,
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	ai.SetDifficulty(DifficultyExpert)
	if ai.GetDifficulty() != DifficultyExpert {
		t.Errorf("Expected difficulty %v, got %v", DifficultyExpert, ai.GetDifficulty())
	}
}

func TestLLMAIEngine_GetProvider(t *testing.T) {
	config := LLMConfig{
		Provider: ProviderXAI,
		APIKey:   "test-key",
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	if ai.GetProvider() != ProviderXAI {
		t.Errorf("Expected provider %v, got %v", ProviderXAI, ai.GetProvider())
	}
}

func TestLLMAIEngine_parseMoveFromResponse(t *testing.T) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	game := engine.NewGame()

	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "valid move",
			response: "e2e4",
			wantErr:  false,
		},
		{
			name:     "move with prefix",
			response: "Move: g1f3",
			wantErr:  false,
		},
		{
			name:     "move with suffix",
			response: "d2d4!",
			wantErr:  false,
		},
		{
			name:     "quoted move",
			response: "\"e2e4\"",
			wantErr:  false,
		},
		{
			name:     "random fallback",
			response: "random",
			wantErr:  false,
		},
		{
			name:     "invalid move",
			response: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			move, err := ai.parseMoveFromResponse(tt.response, game)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMoveFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && move.String() == "" {
				t.Error("parseMoveFromResponse() returned empty move")
			}
		})
	}
}

func TestLLMAIEngine_generateChessPrompt(t *testing.T) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	game := engine.NewGame()
	prompt := ai.generateChessPrompt(game)

	if prompt == "" {
		t.Error("generateChessPrompt() returned empty prompt")
	}

	// Check that prompt contains board representation
	if !contains(prompt, "a b c d e f g h") {
		t.Error("generateChessPrompt() doesn't contain board coordinates")
	}

	// Check that prompt contains color information
	if !contains(prompt, "Active color:") {
		t.Error("generateChessPrompt() doesn't contain active color")
	}
}

func TestLLMAIEngine_boardToString(t *testing.T) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	game := engine.NewGame()
	board := game.Board()
	boardStr := ai.boardToString(board)

	if boardStr == "" {
		t.Error("boardToString() returned empty string")
	}

	// Check for standard chess pieces
	if !contains(boardStr, "r") || !contains(boardStr, "R") {
		t.Error("boardToString() doesn't contain rooks")
	}

	if !contains(boardStr, "k") || !contains(boardStr, "K") {
		t.Error("boardToString() doesn't contain kings")
	}
}

func TestLLMAIEngine_getTemperatureForDifficulty(t *testing.T) {
	config := LLMConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		Difficulty: DifficultyBeginner,
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	// Test that beginner has higher temperature (more randomness)
	beginnerTemp := ai.getTemperatureForDifficulty()

	ai.SetDifficulty(DifficultyExpert)
	expertTemp := ai.getTemperatureForDifficulty()

	if beginnerTemp <= expertTemp {
		t.Errorf("Expected beginner temperature (%f) to be higher than expert temperature (%f)", beginnerTemp, expertTemp)
	}
}

func TestNewLLMAIFromEnv(t *testing.T) {
	// Test with invalid provider
	_, err := NewLLMAIFromEnv("invalid", DifficultyMedium)
	if err == nil {
		t.Error("Expected error for invalid provider")
	}

	// Test with valid provider but no env var (should fail)
	_, err = NewLLMAIFromEnv("openai", DifficultyMedium)
	if err == nil {
		t.Error("Expected error when OPENAI_API_KEY is not set")
	}
}

func TestLLMAIEngine_addToContext(t *testing.T) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	// Add some messages
	ai.addToContext("user", "Hello")
	ai.addToContext("assistant", "Hi there")

	if len(ai.context) != 2 {
		t.Errorf("Expected 2 messages in context, got %d", len(ai.context))
	}

	// Test context trimming by adding many messages
	for i := 0; i < 15; i++ {
		ai.addToContext("user", "test message")
	}

	if len(ai.context) > 10 {
		t.Errorf("Expected context to be trimmed to 10 messages, got %d", len(ai.context))
	}
}

// Benchmark tests
func BenchmarkLLMAIEngine_generateChessPrompt(b *testing.B) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, _ := NewLLMAIEngine(config)
	game := engine.NewGame()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ai.generateChessPrompt(game)
	}
}

func BenchmarkLLMAIEngine_boardToString(b *testing.B) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, _ := NewLLMAIEngine(config)
	game := engine.NewGame()
	board := game.Board()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ai.boardToString(board)
	}
}

// Test that LLMAIEngine implements the Engine interface
func TestLLMAIEngine_ImplementsEngine(t *testing.T) {
	var _ Engine = (*LLMAIEngine)(nil)
}

// Test context timeout
func TestLLMAIEngine_ContextTimeout(t *testing.T) {
	config := LLMConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	ai, err := NewLLMAIEngine(config)
	if err != nil {
		t.Fatalf("Failed to create AI engine: %v", err)
	}

	game := engine.NewGame()

	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should timeout and fallback to RandomAI, but since we pass the same
	// timed-out context, the RandomAI will also fail with context deadline exceeded
	_, err = ai.GetBestMove(ctx, game)
	// We expect this to fail because the context is expired
	if err == nil {
		t.Error("Expected GetBestMove to fail due to context timeout")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
