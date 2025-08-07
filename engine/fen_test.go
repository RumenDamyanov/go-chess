package engine

import (
	"testing"
)

func TestToFEN(t *testing.T) {
	// Test starting position FEN
	game := NewGame()
	fen := game.ToFEN()
	expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	if fen != expected {
		t.Errorf("Expected FEN %s, got %s", expected, fen)
	}
}

func TestToFENAfterMove(t *testing.T) {
	// Test FEN after a move
	game := NewGame()

	// Make a move: e2e4
	move, err := game.ParseMove("e2e4")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	err = game.MakeMove(move)
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

	fen := game.ToFEN()
	expected := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"

	if fen != expected {
		t.Errorf("Expected FEN %s, got %s", expected, fen)
	}
}
