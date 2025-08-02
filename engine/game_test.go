package engine

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	game := NewGame()

	if game == nil {
		t.Fatal("Game should not be nil")
	}
	if game.ActiveColor() != White {
		t.Error("Expected white to move first")
	}
	if game.Status() != InProgress {
		t.Error("Expected game status to be in progress")
	}
	if game.MoveCount() != 1 {
		t.Error("Expected move count to be 1")
	}
	if len(game.MoveHistory()) != 0 {
		t.Error("Expected empty move history")
	}

	// Test initial board state
	board := game.Board()
	if board.GetPiece(E1) != (Piece{Type: King, Color: White}) {
		t.Error("Expected white king on e1")
	}
	if board.GetPiece(E8) != (Piece{Type: King, Color: Black}) {
		t.Error("Expected black king on e8")
	}
}

func TestParseMove(t *testing.T) {
	game := NewGame()

	// Test basic move
	move, err := game.ParseMove("e2e4")
	if err != nil {
		t.Errorf("Unexpected error parsing e2e4: %v", err)
	}
	if move.From != E2 || move.To != E4 {
		t.Error("Incorrect move squares")
	}

	// Test invalid move
	_, err = game.ParseMove("invalid")
	if err == nil {
		t.Error("Expected error for invalid move")
	}

	// Test castling
	move, err = game.ParseMove("O-O")
	if err != nil {
		t.Errorf("Unexpected error parsing O-O: %v", err)
	}
	if move.Type != Castling {
		t.Error("Expected castling move type")
	}
}

func TestMoveString(t *testing.T) {
	tests := []struct {
		move     Move
		expected string
	}{
		{
			Move{From: E2, To: E4, Type: Normal},
			"e2e4",
		},
		{
			Move{From: E7, To: E8, Type: Promotion, Promotion: Queen},
			"e7e8Q",
		},
		{
			Move{From: E1, To: G1, Type: Castling},
			"O-O",
		},
		{
			Move{From: E1, To: C1, Type: Castling},
			"O-O-O",
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.move.String()
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestGameStatusString(t *testing.T) {
	if InProgress.String() != "in_progress" {
		t.Error("Incorrect status string for InProgress")
	}
	if WhiteWins.String() != "white_wins" {
		t.Error("Incorrect status string for WhiteWins")
	}
}

func TestBasicMove(t *testing.T) {
	game := NewGame()

	// Test basic pawn move
	move, err := game.ParseMove("e2e4")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	err = game.MakeMove(move)
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

	// Check that the move was made
	board := game.Board()
	if board.GetPiece(E2) != (Piece{Type: Empty}) {
		t.Error("Expected empty square on e2")
	}
	if board.GetPiece(E4) != (Piece{Type: Pawn, Color: White}) {
		t.Error("Expected white pawn on e4")
	}

	// Check game state
	if game.ActiveColor() != Black {
		t.Error("Expected black to move after white's move")
	}
	if len(game.MoveHistory()) != 1 {
		t.Error("Expected one move in history")
	}
}

// Benchmark tests
func BenchmarkNewGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewGame()
	}
}

func BenchmarkParseMove(b *testing.B) {
	game := NewGame()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = game.ParseMove("e2e4")
	}
}
