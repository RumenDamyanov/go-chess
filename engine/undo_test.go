package engine

import "testing"

// TestUndoMoveSimple ensures a single move can be undone and state restored.
func TestUndoMoveSimple(t *testing.T) {
	g := NewGame()
	mv, err := g.ParseMove("e2e4")
	if err != nil {
		t.Fatalf("parse move: %v", err)
	}
	if err := g.MakeMove(mv); err != nil {
		t.Fatalf("make move: %v", err)
	}
	if len(g.MoveHistory()) != 1 {
		t.Fatalf("expected 1 move, got %d", len(g.MoveHistory()))
	}
	if _, err := g.UndoMove(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if len(g.MoveHistory()) != 0 {
		t.Fatalf("expected 0 moves after undo, got %d", len(g.MoveHistory()))
	}
	// After undo, white pawn should be back on e2 and e4 empty.
	e2, _ := SquareFromString("e2")
	e4, _ := SquareFromString("e4")
	if p := g.Board().GetPiece(e2); p.Type != Pawn || p.Color != White {
		t.Fatalf("expected white pawn on e2 after undo")
	}
	if p := g.Board().GetPiece(e4); !p.IsEmpty() {
		t.Fatalf("expected e4 empty after undo")
	}
	if g.ActiveColor() != White {
		t.Fatalf("expected active color white after undo, got %v", g.ActiveColor())
	}
}

// TestUndoMoveMultiple verifies multiple undo operations restore earlier states.
func TestUndoMoveMultiple(t *testing.T) {
	g := NewGame()
	seq := []string{"e2e4", "e7e5"}
	for _, s := range seq {
		mv, err := g.ParseMove(s)
		if err != nil {
			t.Fatalf("parse %s: %v", s, err)
		}
		if err := g.MakeMove(mv); err != nil {
			t.Fatalf("make %s: %v", s, err)
		}
	}
	if len(g.MoveHistory()) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(g.MoveHistory()))
	}
	// First undo (removes black's move)
	if _, err := g.UndoMove(); err != nil {
		t.Fatalf("undo 1: %v", err)
	}
	if len(g.MoveHistory()) != 1 {
		t.Fatalf("expected 1 move after first undo, got %d", len(g.MoveHistory()))
	}
	// Position should be after e2e4
	e4, _ := SquareFromString("e4")
	e7, _ := SquareFromString("e7")
	if p := g.Board().GetPiece(e4); p.Type != Pawn || p.Color != White {
		t.Fatalf("expected white pawn on e4 after first undo")
	}
	if p := g.Board().GetPiece(e7); p.Type != Pawn || p.Color != Black {
		t.Fatalf("expected black pawn on e7 after first undo")
	}
	if g.ActiveColor() != Black { // black to move after e2e4
		t.Fatalf("expected active color black after first undo, got %v", g.ActiveColor())
	}
	// Second undo (removes white's move)
	if _, err := g.UndoMove(); err != nil {
		t.Fatalf("undo 2: %v", err)
	}
	if len(g.MoveHistory()) != 0 {
		t.Fatalf("expected 0 moves after second undo, got %d", len(g.MoveHistory()))
	}
	e2b, _ := SquareFromString("e2")
	e4b, _ := SquareFromString("e4")
	if p := g.Board().GetPiece(e2b); p.Type != Pawn || p.Color != White {
		t.Fatalf("expected white pawn on e2 after second undo")
	}
	if p := g.Board().GetPiece(e4b); !p.IsEmpty() {
		t.Fatalf("expected e4 empty after second undo")
	}
	if g.ActiveColor() != White { // initial position
		t.Fatalf("expected active color white after second undo, got %v", g.ActiveColor())
	}
}

// TestUndoMoveEmpty ensures undo on a fresh game returns an error.
func TestUndoMoveEmpty(t *testing.T) {
	g := NewGame()
	if _, err := g.UndoMove(); err == nil {
		t.Fatalf("expected error undoing with no moves")
	}
}
