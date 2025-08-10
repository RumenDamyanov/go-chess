package ai

import (
	"context"
	"testing"
	"time"

	"go.rumenx.com/chess/engine"
)

// Test internal helper methods of MinimaxAI that previously had 0% coverage.
func TestMinimaxAI_isPathClear(t *testing.T) {
	ai := NewMinimaxAI(DifficultyEasy)

	game := engine.NewGame()
	// In the starting position, path between a1 (white rook) and a8 (black rook) is blocked by pawns.
	if ai.isPathClear(game, engine.A1, engine.A8) {
		t.Errorf("expected path a1->a8 to be blocked in starting position")
	}

	// Use a simplified FEN with only rooks and kings to test a clear vertical file.
	// r7k is invalid (9 files). Use r6k (r + 6 empty + k = 8) and R6K for rank1.
	fen := "r6k/8/8/8/8/8/8/R6K w - - 0 1"
	custom := engine.NewGame()
	if err := custom.ParseFEN(fen); err != nil {
		t.Fatalf("failed to load FEN: %v", err)
	}
	if !ai.isPathClear(custom, engine.A1, engine.A8) {
		t.Errorf("expected path a1->a8 to be clear in custom position")
	}
}

func TestMinimaxAI_findCheckEscapeMovesFiltersIllegal(t *testing.T) {
	ai := NewMinimaxAI(DifficultyMedium)
	game := engine.NewGame()

	legal := ai.GenerateLegalMoves(game)
	if len(legal) == 0 {
		t.Fatalf("expected legal moves in starting position")
	}

	// Construct an obviously illegal move (rook jumps over pieces) and include it in the slice.
	illegal := engine.Move{From: engine.A1, To: engine.A4, Type: engine.Normal, Piece: game.Board().GetPiece(engine.A1)}
	mixed := append([]engine.Move{illegal}, legal[0])

	filtered := ai.findCheckEscapeMoves(game, mixed, engine.White)
	// The illegal move should be discarded, leaving only the legal move we added.
	if len(filtered) != 1 {
		t.Fatalf("expected 1 legal escape move, got %d", len(filtered))
	}
	if filtered[0] != legal[0] {
		t.Errorf("unexpected move filtered: got %v want %v", filtered[0], legal[0])
	}
}

// Smoke test GetBestMove still works with context timing (ensures helper usage path executed for coverage).
func TestMinimaxAI_GetBestMoveBasic(t *testing.T) {
	ai := NewMinimaxAI(DifficultyEasy)
	game := engine.NewGame()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mv, err := ai.GetBestMove(ctx, game)
	if err != nil {
		t.Fatalf("GetBestMove returned error: %v", err)
	}
	if mv.From == mv.To {
		t.Errorf("expected a non-null move, got %+v", mv)
	}
}
