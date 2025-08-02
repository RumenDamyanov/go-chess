package engine

import (
	"testing"
)

func TestSquareFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected Square
		hasError bool
	}{
		{"a1", A1, false},
		{"h8", H8, false},
		{"e4", E4, false},
		{"d5", D5, false},
		{"", 0, true},
		{"z9", 0, true},
		{"aa", 0, true},
		{"11", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := SquareFromString(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %v, got %v for input %s", test.expected, result, test.input)
				}
			}
		})
	}
}

func TestSquareString(t *testing.T) {
	tests := []struct {
		square   Square
		expected string
	}{
		{A1, "a1"},
		{H8, "h8"},
		{E4, "e4"},
		{D5, "d5"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.square.String()
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestNewBoard(t *testing.T) {
	board := NewBoard()

	if board == nil {
		t.Fatal("Board should not be nil")
	}

	// Test white pieces
	if board.GetPiece(A1) != (Piece{Type: Rook, Color: White}) {
		t.Error("Expected white rook on a1")
	}
	if board.GetPiece(E1) != (Piece{Type: King, Color: White}) {
		t.Error("Expected white king on e1")
	}

	// Test black pieces
	if board.GetPiece(A8) != (Piece{Type: Rook, Color: Black}) {
		t.Error("Expected black rook on a8")
	}
	if board.GetPiece(E8) != (Piece{Type: King, Color: Black}) {
		t.Error("Expected black king on e8")
	}

	// Test empty squares
	if board.GetPiece(E4) != (Piece{Type: Empty}) {
		t.Error("Expected empty square on e4")
	}
}

func TestColorString(t *testing.T) {
	if White.String() != "white" {
		t.Errorf("Expected 'white', got %s", White.String())
	}
	if Black.String() != "black" {
		t.Errorf("Expected 'black', got %s", Black.String())
	}
}

func TestPieceString(t *testing.T) {
	tests := []struct {
		piece    Piece
		expected string
	}{
		{Piece{Type: Empty}, "."},
		{Piece{Type: King, Color: White}, "K"},
		{Piece{Type: King, Color: Black}, "k"},
		{Piece{Type: Pawn, Color: White}, "P"},
		{Piece{Type: Pawn, Color: Black}, "p"},
	}

	for _, test := range tests {
		result := test.piece.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

// Benchmark tests
func BenchmarkSquareFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SquareFromString("e4")
	}
}

func BenchmarkBoardCopy(b *testing.B) {
	board := NewBoard()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = board.Copy()
	}
}
