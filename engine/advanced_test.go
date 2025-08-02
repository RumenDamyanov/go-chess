package engine

import (
	"testing"
)

// TestCastling tests castling functionality
func TestCastling(t *testing.T) {
	game := NewGame()

	// Clear pieces between king and rook for white kingside castling
	// Move knights and bishops out of the way
	game.ParseMove("Ng1f3")
	game.MakeMove(Move{From: G1, To: F3, Type: Normal, Piece: Piece{Type: Knight, Color: White}})
	game.ParseMove("Bf1e2")
	game.MakeMove(Move{From: F1, To: E2, Type: Normal, Piece: Piece{Type: Bishop, Color: White}})

	// Test castling move parsing
	move, err := game.ParseMove("O-O")
	if err != nil {
		t.Errorf("Failed to parse castling move: %v", err)
	}
	if move.Type != Castling {
		t.Error("Expected castling move type")
	}

	// Test alternative castling notation
	move2, err := game.ParseMove("0-0")
	if err != nil {
		t.Errorf("Failed to parse alternative castling notation: %v", err)
	}
	if move2.Type != Castling {
		t.Error("Expected castling move type for alternative notation")
	}

	// Test queenside castling
	move3, err := game.ParseMove("O-O-O")
	if err != nil {
		t.Errorf("Failed to parse queenside castling: %v", err)
	}
	if move3.Type != Castling {
		t.Error("Expected castling move type for queenside")
	}
}

// TestPromotion tests pawn promotion
func TestPromotion(t *testing.T) {
	game := NewGame()

	// Move white pawn to 7th rank manually for testing
	game.board.SetPiece(E7, Piece{Type: Pawn, Color: White})
	game.board.SetPiece(E2, Piece{Type: Empty})

	// Test promotion move parsing
	move, err := game.ParseMove("e7e8Q")
	if err != nil {
		t.Errorf("Failed to parse promotion move: %v", err)
	}
	if move.Type != Promotion {
		t.Error("Expected promotion move type")
	}
	if move.Promotion != Queen {
		t.Error("Expected queen promotion")
	}

	// Test other promotion pieces
	pieces := []struct {
		notation string
		piece    PieceType
	}{
		{"e7e8R", Rook},
		{"e7e8B", Bishop},
		{"e7e8N", Knight},
	}

	for _, p := range pieces {
		move, err := game.ParseMove(p.notation)
		if err != nil {
			t.Errorf("Failed to parse %s promotion: %v", p.notation, err)
		}
		if move.Promotion != p.piece {
			t.Errorf("Expected %v promotion, got %v", p.piece, move.Promotion)
		}
	}
}

// TestEnPassant tests en passant capture
func TestEnPassant(t *testing.T) {
	game := NewGame()

	// Set up en passant scenario
	game.board.SetPiece(E5, Piece{Type: Pawn, Color: White})
	game.board.SetPiece(D5, Piece{Type: Pawn, Color: Black})
	game.enPassantSquare = D6
	game.activeColor = White

	// Test en passant move
	move := Move{
		From:     E5,
		To:       D6,
		Type:     EnPassant,
		Piece:    Piece{Type: Pawn, Color: White},
		Captured: Piece{Type: Pawn, Color: Black},
	}

	if !game.IsLegalMove(move) {
		t.Error("En passant move should be legal")
	}
}

// TestPieceMovement tests different piece movement validation
func TestPieceMovement(t *testing.T) {
	game := NewGame()

	tests := []struct {
		name    string
		from    Square
		to      Square
		piece   Piece
		isLegal bool
	}{
		// Pawn moves
		{"pawn forward one", E2, E3, Piece{Type: Pawn, Color: White}, true},
		{"pawn forward two", E2, E4, Piece{Type: Pawn, Color: White}, true},
		{"pawn invalid sideways", E2, F2, Piece{Type: Pawn, Color: White}, false},

		// Knight moves
		{"knight L-shape", B1, C3, Piece{Type: Knight, Color: White}, true},
		{"knight invalid straight", B1, B3, Piece{Type: Knight, Color: White}, false},

		// Rook moves (after clearing path)
		{"rook horizontal", A1, D1, Piece{Type: Rook, Color: White}, false}, // blocked by pieces

		// Bishop moves (after clearing path)
		{"bishop diagonal", C1, F4, Piece{Type: Bishop, Color: White}, false}, // blocked by pawn

		// King moves
		{"king one square", E1, E2, Piece{Type: King, Color: White}, false}, // blocked by pawn
		{"king invalid far", E1, E3, Piece{Type: King, Color: White}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			move := Move{
				From:  test.from,
				To:    test.to,
				Type:  Normal,
				Piece: test.piece,
			}

			result := game.isPseudoLegalMove(move)
			if result != test.isLegal {
				t.Errorf("Expected %v for %s, got %v", test.isLegal, test.name, result)
			}
		})
	}
}

// TestGameStatus tests game status detection
func TestGameStatus(t *testing.T) {
	// Test all game status string representations
	statuses := []struct {
		status   GameStatus
		expected string
	}{
		{InProgress, "in_progress"},
		{WhiteWins, "white_wins"},
		{BlackWins, "black_wins"},
		{Draw, "draw"},
	}

	for _, s := range statuses {
		result := s.status.String()
		if result != s.expected {
			t.Errorf("Expected %s for status %v, got %s", s.expected, s.status, result)
		}
	}
}

// TestMoveType tests move type string representations
func TestMoveType(t *testing.T) {
	moveTypes := []struct {
		moveType MoveType
		expected string
	}{
		{Normal, "normal"},
		{Capture, "capture"},
		{Castling, "castling"},
		{EnPassant, "en_passant"},
		{Promotion, "promotion"},
	}

	for _, mt := range moveTypes {
		result := mt.moveType.String()
		if result != mt.expected {
			t.Errorf("Expected %s for move type %v, got %s", mt.expected, mt.moveType, result)
		}
	}
}

// TestInvalidMoves tests various invalid move scenarios
func TestInvalidMoves(t *testing.T) {
	game := NewGame()

	invalidMoves := []string{
		"",        // empty
		"e2",      // too short
		"e2e",     // too short
		"z9z9",    // invalid squares
		"e2e9",    // invalid target square
		"e9e2",    // invalid source square
		"invalid", // completely invalid
	}

	for _, moveStr := range invalidMoves {
		_, err := game.ParseMove(moveStr)
		if err == nil {
			t.Errorf("Expected error for invalid move: %s", moveStr)
		}
	}
}

// TestCastlingRights tests castling rights management
func TestCastlingRights(t *testing.T) {
	game := NewGame()

	// Initially all castling rights should be available
	if !game.castlingRights.WhiteKingside {
		t.Error("Expected white kingside castling to be available")
	}
	if !game.castlingRights.WhiteQueenside {
		t.Error("Expected white queenside castling to be available")
	}
	if !game.castlingRights.BlackKingside {
		t.Error("Expected black kingside castling to be available")
	}
	if !game.castlingRights.BlackQueenside {
		t.Error("Expected black queenside castling to be available")
	}
}

// TestBoardCopy tests that board copies are independent
func TestBoardCopy(t *testing.T) {
	game := NewGame()
	originalBoard := game.Board()

	// Make a move in the game
	move, _ := game.ParseMove("e2e4")
	game.MakeMove(move)

	// Original board copy should be unchanged
	if originalBoard.GetPiece(E2) != (Piece{Type: Pawn, Color: White}) {
		t.Error("Board copy should be independent of original")
	}
	if originalBoard.GetPiece(E4) != (Piece{Type: Empty}) {
		t.Error("Board copy should not reflect moves made after copying")
	}
}

// TestMoveHistory tests move history functionality
func TestMoveHistory(t *testing.T) {
	game := NewGame()

	// Initially no moves
	if len(game.MoveHistory()) != 0 {
		t.Error("Expected empty move history initially")
	}

	// Make some moves
	moves := []string{"e2e4", "e7e5", "b1c3"}

	for i, moveStr := range moves {
		move, err := game.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}

		err = game.MakeMove(move)
		if err != nil {
			t.Fatalf("Failed to make move %s: %v", moveStr, err)
		}

		history := game.MoveHistory()
		if len(history) != i+1 {
			t.Errorf("Expected %d moves in history, got %d", i+1, len(history))
		}
	}

	// Test that history is a copy (modifications don't affect original)
	history := game.MoveHistory()
	originalLen := len(history)
	history = append(history, Move{}) // This should not affect the game's history

	if len(game.MoveHistory()) != originalLen {
		t.Error("Move history should return a copy, not the original slice")
	}
}
