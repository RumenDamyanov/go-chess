// Package ai provides artificial intelligence implementations for the chess engine.
// It includes various AI algorithms with different difficulty levels and playing styles.
package ai

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/rumendamyanov/go-chess/engine"
)

// Difficulty represents the difficulty level of an AI opponent.
type Difficulty int

const (
	// DifficultyBeginner represents the easiest AI level.
	DifficultyBeginner Difficulty = iota
	// DifficultyEasy represents an easy AI level.
	DifficultyEasy
	// DifficultyMedium represents a medium AI level.
	DifficultyMedium
	// DifficultyHard represents a hard AI level.
	DifficultyHard
	// DifficultyExpert represents the hardest AI level.
	DifficultyExpert
)

// String returns the string representation of a difficulty level.
func (d Difficulty) String() string {
	switch d {
	case DifficultyBeginner:
		return "beginner"
	case DifficultyEasy:
		return "easy"
	case DifficultyMedium:
		return "medium"
	case DifficultyHard:
		return "hard"
	case DifficultyExpert:
		return "expert"
	default:
		return "unknown"
	}
}

// Engine represents an AI chess engine interface.
type Engine interface {
	// GetBestMove returns the best move for the current position.
	GetBestMove(ctx context.Context, game *engine.Game) (engine.Move, error)

	// GetDifficulty returns the difficulty level of the AI.
	GetDifficulty() Difficulty

	// SetDifficulty sets the difficulty level of the AI.
	SetDifficulty(difficulty Difficulty)
}

// RandomAI implements a simple random move AI.
type RandomAI struct {
	difficulty Difficulty
	rng        *rand.Rand
}

// NewRandomAI creates a new random AI with beginner difficulty.
func NewRandomAI() *RandomAI {
	return &RandomAI{
		difficulty: DifficultyBeginner,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetBestMove returns a random legal move.
func (ai *RandomAI) GetBestMove(ctx context.Context, game *engine.Game) (engine.Move, error) {
	moves := ai.generateLegalMoves(game)
	if len(moves) == 0 {
		return engine.Move{}, errors.New("no legal moves available")
	}

	// Add some delay based on difficulty to simulate thinking
	thinkTime := time.Duration(ai.rng.Intn(1000)) * time.Millisecond
	select {
	case <-time.After(thinkTime):
	case <-ctx.Done():
		return engine.Move{}, ctx.Err()
	}

	return moves[ai.rng.Intn(len(moves))], nil
}

// GetDifficulty returns the current difficulty level.
func (ai *RandomAI) GetDifficulty() Difficulty {
	return ai.difficulty
}

// SetDifficulty sets the difficulty level.
func (ai *RandomAI) SetDifficulty(difficulty Difficulty) {
	ai.difficulty = difficulty
}

// generateLegalMoves generates all legal moves for the current position.
func (ai *RandomAI) generateLegalMoves(game *engine.Game) []engine.Move {
	var moves []engine.Move
	board := game.Board()
	activeColor := game.ActiveColor()

	// Iterate through all squares to find pieces of the active color
	for sq := engine.Square(0); sq < 64; sq++ {
		piece := board.GetPiece(sq)
		if piece.IsEmpty() || piece.Color != activeColor {
			continue
		}

		// Generate possible moves for this piece
		pieceMoves := ai.generatePieceMovesI(game, sq, piece)
		for _, move := range pieceMoves {
			if game.IsLegalMove(move) {
				moves = append(moves, move)
			}
		}
	}

	return moves
}

// generatePieceMovesI generates possible moves for a piece at a given square.
func (ai *RandomAI) generatePieceMovesI(game *engine.Game, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move
	board := game.Board()

	switch piece.Type {
	case engine.Pawn:
		moves = append(moves, ai.generatePawnMoves(board, from, piece)...)
	case engine.Rook:
		moves = append(moves, ai.generateRookMoves(board, from, piece)...)
	case engine.Knight:
		moves = append(moves, ai.generateKnightMoves(board, from, piece)...)
	case engine.Bishop:
		moves = append(moves, ai.generateBishopMoves(board, from, piece)...)
	case engine.Queen:
		moves = append(moves, ai.generateQueenMoves(board, from, piece)...)
	case engine.King:
		moves = append(moves, ai.generateKingMoves(board, from, piece)...)
	}

	return moves
}

// generatePawnMoves generates possible pawn moves.
func (ai *RandomAI) generatePawnMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move
	direction := 1
	if piece.Color == engine.Black {
		direction = -1
	}

	// Forward move
	to := engine.Square(int(from) + direction*8)
	if to >= 0 && to < 64 && board.GetPiece(to).IsEmpty() {
		moves = append(moves, engine.Move{
			From:  from,
			To:    to,
			Type:  engine.Normal,
			Piece: piece,
		})

		// Double move from starting position
		if (piece.Color == engine.White && from.Rank() == 1) ||
			(piece.Color == engine.Black && from.Rank() == 6) {
			to2 := engine.Square(int(to) + direction*8)
			if to2 >= 0 && to2 < 64 && board.GetPiece(to2).IsEmpty() {
				moves = append(moves, engine.Move{
					From:  from,
					To:    to2,
					Type:  engine.Normal,
					Piece: piece,
				})
			}
		}
	}

	// Diagonal captures
	for _, fileOffset := range []int{-1, 1} {
		to := engine.Square(int(from) + direction*8 + fileOffset)
		if to >= 0 && to < 64 && abs(to.File()-from.File()) == 1 {
			target := board.GetPiece(to)
			if !target.IsEmpty() && target.Color != piece.Color {
				moves = append(moves, engine.Move{
					From:     from,
					To:       to,
					Type:     engine.Capture,
					Piece:    piece,
					Captured: target,
				})
			}
		}
	}

	return moves
}

// generateRookMoves generates possible rook moves.
func (ai *RandomAI) generateRookMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move

	// Horizontal and vertical directions
	directions := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			to := engine.Square(int(from) + i*(dir[0]+dir[1]*8))
			if to < 0 || to >= 64 || abs(to.File()-from.File()) > 7 || abs(to.Rank()-from.Rank()) > 7 {
				break
			}

			target := board.GetPiece(to)
			if target.IsEmpty() {
				moves = append(moves, engine.Move{
					From:  from,
					To:    to,
					Type:  engine.Normal,
					Piece: piece,
				})
			} else {
				if target.Color != piece.Color {
					moves = append(moves, engine.Move{
						From:     from,
						To:       to,
						Type:     engine.Capture,
						Piece:    piece,
						Captured: target,
					})
				}
				break
			}
		}
	}

	return moves
}

// generateKnightMoves generates possible knight moves.
func (ai *RandomAI) generateKnightMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move

	// Knight move offsets
	offsets := [][]int{{2, 1}, {2, -1}, {-2, 1}, {-2, -1}, {1, 2}, {1, -2}, {-1, 2}, {-1, -2}}

	for _, offset := range offsets {
		to := engine.Square(int(from) + offset[0] + offset[1]*8)
		if to >= 0 && to < 64 && abs(to.File()-from.File()) <= 2 && abs(to.Rank()-from.Rank()) <= 2 {
			target := board.GetPiece(to)
			if target.IsEmpty() {
				moves = append(moves, engine.Move{
					From:  from,
					To:    to,
					Type:  engine.Normal,
					Piece: piece,
				})
			} else if target.Color != piece.Color {
				moves = append(moves, engine.Move{
					From:     from,
					To:       to,
					Type:     engine.Capture,
					Piece:    piece,
					Captured: target,
				})
			}
		}
	}

	return moves
}

// generateBishopMoves generates possible bishop moves.
func (ai *RandomAI) generateBishopMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move

	// Diagonal directions
	directions := [][]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			to := engine.Square(int(from) + i*(dir[0]+dir[1]*8))
			if to < 0 || to >= 64 || abs(to.File()-from.File()) != i || abs(to.Rank()-from.Rank()) != i {
				break
			}

			target := board.GetPiece(to)
			if target.IsEmpty() {
				moves = append(moves, engine.Move{
					From:  from,
					To:    to,
					Type:  engine.Normal,
					Piece: piece,
				})
			} else {
				if target.Color != piece.Color {
					moves = append(moves, engine.Move{
						From:     from,
						To:       to,
						Type:     engine.Capture,
						Piece:    piece,
						Captured: target,
					})
				}
				break
			}
		}
	}

	return moves
}

// generateQueenMoves generates possible queen moves.
func (ai *RandomAI) generateQueenMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move
	moves = append(moves, ai.generateRookMoves(board, from, piece)...)
	moves = append(moves, ai.generateBishopMoves(board, from, piece)...)
	return moves
}

// generateKingMoves generates possible king moves.
func (ai *RandomAI) generateKingMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move

	// King can move one square in any direction
	directions := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	for _, dir := range directions {
		to := engine.Square(int(from) + dir[0] + dir[1]*8)
		if to >= 0 && to < 64 && abs(to.File()-from.File()) <= 1 && abs(to.Rank()-from.Rank()) <= 1 {
			target := board.GetPiece(to)
			if target.IsEmpty() {
				moves = append(moves, engine.Move{
					From:  from,
					To:    to,
					Type:  engine.Normal,
					Piece: piece,
				})
			} else if target.Color != piece.Color {
				moves = append(moves, engine.Move{
					From:     from,
					To:       to,
					Type:     engine.Capture,
					Piece:    piece,
					Captured: target,
				})
			}
		}
	}

	return moves
}

// MinimaxAI implements a minimax AI with alpha-beta pruning.
type MinimaxAI struct {
	difficulty Difficulty
	depth      int
}

// NewMinimaxAI creates a new minimax AI with the specified difficulty.
func NewMinimaxAI(difficulty Difficulty) *MinimaxAI {
	depth := 2
	switch difficulty {
	case DifficultyEasy:
		depth = 2
	case DifficultyMedium:
		depth = 3
	case DifficultyHard:
		depth = 4
	case DifficultyExpert:
		depth = 5
	}

	return &MinimaxAI{
		difficulty: difficulty,
		depth:      depth,
	}
}

// GetBestMove returns the best move using minimax algorithm.
func (ai *MinimaxAI) GetBestMove(ctx context.Context, game *engine.Game) (engine.Move, error) {
	randomAI := NewRandomAI()
	moves := randomAI.generateLegalMoves(game)

	if len(moves) == 0 {
		return engine.Move{}, errors.New("no legal moves available")
	}

	// For now, just return a random move as placeholder
	// TODO: Implement actual minimax algorithm
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Add thinking time based on difficulty
	thinkTime := time.Duration(ai.depth*500) * time.Millisecond
	select {
	case <-time.After(thinkTime):
	case <-ctx.Done():
		return engine.Move{}, ctx.Err()
	}

	return moves[rng.Intn(len(moves))], nil
}

// GetDifficulty returns the current difficulty level.
func (ai *MinimaxAI) GetDifficulty() Difficulty {
	return ai.difficulty
}

// SetDifficulty sets the difficulty level and adjusts search depth.
func (ai *MinimaxAI) SetDifficulty(difficulty Difficulty) {
	ai.difficulty = difficulty
	switch difficulty {
	case DifficultyEasy:
		ai.depth = 2
	case DifficultyMedium:
		ai.depth = 3
	case DifficultyHard:
		ai.depth = 4
	case DifficultyExpert:
		ai.depth = 5
	default:
		ai.depth = 2
	}
}

// Helper function
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
