package ai

import (
	"context"
	"testing"
	"time"

	"github.com/rumendamyanov/go-chess/engine"
)

func TestRandomAI(t *testing.T) {
	ai := NewRandomAI()

	if ai == nil {
		t.Fatal("AI should not be nil")
	}

	if ai.GetDifficulty() != DifficultyBeginner {
		t.Error("Expected beginner difficulty by default")
	}

	// Test difficulty setting
	ai.SetDifficulty(DifficultyMedium)
	if ai.GetDifficulty() != DifficultyMedium {
		t.Error("Failed to set difficulty")
	}
}

func TestMinimaxAI(t *testing.T) {
	ai := NewMinimaxAI(DifficultyMedium)

	if ai == nil {
		t.Fatal("AI should not be nil")
	}

	if ai.GetDifficulty() != DifficultyMedium {
		t.Error("Expected medium difficulty")
	}

	// Test difficulty setting
	ai.SetDifficulty(DifficultyHard)
	if ai.GetDifficulty() != DifficultyHard {
		t.Error("Failed to set difficulty")
	}
}

func TestAIGetBestMove(t *testing.T) {
	game := engine.NewGame()
	ai := NewRandomAI()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := ai.GetBestMove(ctx, game)
	if err != nil {
		t.Fatalf("Failed to get AI move: %v", err)
	}

	// Check that the move is valid
	if !game.IsLegalMove(move) {
		t.Error("AI returned illegal move")
	}
}

func TestAIGetBestMoveWithTimeout(t *testing.T) {
	game := engine.NewGame()
	ai := NewMinimaxAI(DifficultyExpert)

	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := ai.GetBestMove(ctx, game)
	// Should either succeed quickly or timeout
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDifficultyString(t *testing.T) {
	tests := []struct {
		difficulty Difficulty
		expected   string
	}{
		{DifficultyBeginner, "beginner"},
		{DifficultyEasy, "easy"},
		{DifficultyMedium, "medium"},
		{DifficultyHard, "hard"},
		{DifficultyExpert, "expert"},
	}

	for _, test := range tests {
		result := test.difficulty.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGenerateLegalMoves(t *testing.T) {
	game := engine.NewGame()
	ai := NewRandomAI()

	moves := ai.GenerateLegalMoves(game)

	if len(moves) == 0 {
		t.Error("Expected some legal moves in starting position")
	}

	// Check that all generated moves are legal
	for _, move := range moves {
		if !game.IsLegalMove(move) {
			t.Errorf("Generated illegal move: %v", move)
		}
	}
}

// Benchmark tests
func BenchmarkRandomAIGetBestMove(b *testing.B) {
	game := engine.NewGame()
	ai := NewRandomAI()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ai.GetBestMove(ctx, game)
	}
}

func BenchmarkGenerateLegalMoves(b *testing.B) {
	game := engine.NewGame()
	ai := NewRandomAI()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ai.GenerateLegalMoves(game)
	}
}
