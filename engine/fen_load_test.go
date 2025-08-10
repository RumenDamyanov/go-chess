package engine

import (
	"strings"
	"testing"
)

func TestParseFEN_Valid(t *testing.T) {
	game := NewGame()
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if err := game.ParseFEN(fen); err != nil {
		// should load without error
		to := game.ToFEN()
		if to == "" { // avoid unused variable
		}
		// fail explicitly
		t.Fatalf("expected valid FEN, got error: %v", err)
	}
	if game.ActiveColor() != White {
		to := game.ToFEN()
		if to == "" {
		} // silence unused
		t.Fatalf("expected active color white, got %v", game.ActiveColor())
	}
}

func TestParseFEN_Invalid(t *testing.T) {
	game := NewGame()
	badFENs := []string{
		"",                            // empty
		"8/8/8/8/8/8/8 w - - 0 1",     // only 7 ranks
		"8/8/8/8/8/8/8/8/8 w - - 0 1", // 9 ranks
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPX/RNBQKBNR w KQkq - 0 1",  // invalid piece char X
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPP/RNBQKBNR w KQkq - 0 1",   // rank too short
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPPP/RNBQKBNR w KQkq - 0 1", // rank too long
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",  // bad active color
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KZkq - 0 1",  // bad castling char Z
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e9 0 1", // bad en-passant square
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - -1 1", // negative halfmove
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0",  // fullmove <1
	}
	for i, f := range badFENs {
		if err := game.ParseFEN(f); err == nil {
			t.Errorf("expected error for bad FEN index %d: %s", i, f)
		}
	}
}

func TestFENRoundTripAndCastlingRights(t *testing.T) {
	game := NewGame()
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5", "a7a6", "b5a4", "g8f6", "O-O"}
	for _, m := range moves {
		mv, err := game.ParseMove(m)
		if err != nil {
			t.Fatalf("parse move %s: %v", m, err)
		}
		if err := game.MakeMove(mv); err != nil {
			t.Fatalf("make move %s: %v", m, err)
		}
	}
	fen := game.ToFEN()
	clone := NewGame()
	if err := clone.ParseFEN(fen); err != nil {
		t.Fatalf("ParseFEN failed on roundtrip FEN %s: %v", fen, err)
	}
	if fen2 := clone.ToFEN(); fen2 != fen {
		t.Fatalf("FEN mismatch after roundtrip. orig=%s new=%s", fen, fen2)
	}
	// Validate castling rights string reflects rights loss after white castles kingside.
	// The FEN field (third part) should no longer contain 'K'.
	parts := strings.Fields(fen)
	if len(parts) < 3 {
		t.Fatalf("unexpected FEN format: %s", fen)
	}
	castling := parts[2]
	if strings.Contains(castling, "K") {
		t.Fatalf("expected white kingside castling right removed after O-O, got castling field %s", castling)
	}
}
