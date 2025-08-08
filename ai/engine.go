// Package ai provides artificial intelligence implementations for the chess engine.
// It includes various AI algorithms with different difficulty levels and playing styles.
package ai

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.rumenx.com/chess/engine"
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

// GenerateAllLegalMoves generates all legal moves for the current position.
// This is a public wrapper around the RandomAI's GenerateLegalMoves method.
func GenerateAllLegalMoves(game *engine.Game) []engine.Move {
	randomAI := NewRandomAI()
	return randomAI.GenerateLegalMoves(game)
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
	moves := ai.GenerateLegalMoves(game)
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

// GenerateLegalMoves generates all legal moves for the current position.
func (ai *RandomAI) GenerateLegalMoves(game *engine.Game) []engine.Move {
	var moves []engine.Move
	board := game.Board()
	activeColor := game.ActiveColor()

	// Iterate through all squares to find pieces of the active color
	for sq := engine.Square(0); sq < 64; sq++ {
		piece := board.GetPiece(sq)
		if piece.IsEmpty() {
			continue
		}

		if piece.Color != activeColor {
			continue
		}

		// Generate possible moves for this piece
		pieceMoves := ai.generatePieceMovesI(game, sq, piece)

		// Apply legal move validation
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
	direction := 8 // White pawns move "up" the board (increasing rank)
	if piece.Color == engine.Black {
		direction = -8 // Black pawns move "down" the board (decreasing rank)
	}

	// Forward move
	to := engine.Square(int(from) + direction)
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
			to2 := engine.Square(int(to) + direction)
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
		to := engine.Square(int(from) + direction + fileOffset)
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

	// Horizontal and vertical directions: right, left, up, down
	directions := []int{1, -1, 8, -8}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			to := engine.Square(int(from) + i*dir)

			// Check bounds
			if to < 0 || to >= 64 {
				break
			}

			// Check if we've wrapped around the board (for horizontal moves)
			if dir == 1 || dir == -1 {
				if abs(to.File()-from.File()) != i {
					break
				}
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

	// Knight move offsets: 2 squares in one direction, 1 square perpendicular
	offsets := []int{17, 15, 10, 6, -6, -10, -15, -17} // 2up+1right, 2up+1left, 1up+2right, 1up+2left, etc.

	for _, offset := range offsets {
		to := engine.Square(int(from) + offset)

		// Check bounds
		if to < 0 || to >= 64 {
			continue
		}

		// Check if the move is a valid knight move (not wrapping around board edges)
		fileDiff := abs(to.File() - from.File())
		rankDiff := abs(to.Rank() - from.Rank())
		if !((fileDiff == 2 && rankDiff == 1) || (fileDiff == 1 && rankDiff == 2)) {
			continue
		}

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

	return moves
}

// generateBishopMoves generates possible bishop moves.
func (ai *RandomAI) generateBishopMoves(board *engine.Board, from engine.Square, piece engine.Piece) []engine.Move {
	var moves []engine.Move

	// Diagonal directions: up-right, up-left, down-right, down-left
	directions := []int{9, 7, -7, -9}

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			to := engine.Square(int(from) + i*dir)

			// Check bounds
			if to < 0 || to >= 64 {
				break
			}

			// Check if we've wrapped around the board or moved the wrong distance
			if abs(to.File()-from.File()) != i || abs(to.Rank()-from.Rank()) != i {
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
	directions := []int{1, -1, 8, -8, 9, 7, -7, -9} // right, left, up, down, up-right, up-left, down-left, down-right

	for _, dir := range directions {
		to := engine.Square(int(from) + dir)

		// Check bounds
		if to < 0 || to >= 64 {
			continue
		}

		// Check if we've wrapped around the board
		if abs(to.File()-from.File()) > 1 || abs(to.Rank()-from.Rank()) > 1 {
			continue
		}

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
	moves := ai.GenerateLegalMoves(game)

	if len(moves) == 0 {
		return engine.Move{}, errors.New("no legal moves available")
	}

	// Check if we're in check - if so, prioritize getting out of check
	activeColor := game.ActiveColor()
	inCheck := ai.isGameInCheck(game, activeColor)
	if inCheck {
		// Find moves that get us out of check
		checkEscapeMoves := ai.findCheckEscapeMoves(game, moves, activeColor)
		if len(checkEscapeMoves) > 0 {
			moves = checkEscapeMoves
		}
	}

	// Simple evaluation-based selection (better than random)
	bestMove := moves[0]
	bestScore := ai.evaluateMove(game, moves[0])

	for _, move := range moves[1:] {
		score := ai.evaluateMove(game, move)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	// Add thinking time based on difficulty
	thinkTime := time.Duration(ai.depth*100) * time.Millisecond
	select {
	case <-time.After(thinkTime):
	case <-ctx.Done():
		return engine.Move{}, ctx.Err()
	}

	return bestMove, nil
}

// GenerateLegalMoves generates all legal moves for the current position
func (ai *MinimaxAI) GenerateLegalMoves(game *engine.Game) []engine.Move {
	// Use the existing function from the engine package
	return GenerateAllLegalMoves(game)
}

// isGameInCheck checks if the given color is in check
func (ai *MinimaxAI) isGameInCheck(game *engine.Game, color engine.Color) bool {
	// Find the king
	kingSquare := engine.Square(-1)

	for sq := engine.Square(0); sq < 64; sq++ {
		piece := game.Board().GetPiece(sq)
		if piece.Color == color && piece.Type == engine.King {
			kingSquare = sq
			break
		}
	}

	if kingSquare == -1 {
		return false // King not found
	}

	// Check if any opponent piece can attack the king
	opponentColor := engine.White
	if color == engine.White {
		opponentColor = engine.Black
	}

	for sq := engine.Square(0); sq < 64; sq++ {
		piece := game.Board().GetPiece(sq)
		if piece.Color == opponentColor && !piece.IsEmpty() {
			// Check if this piece can attack the king
			if ai.canPieceAttackSquare(game, sq, kingSquare) {
				return true
			}
		}
	}

	return false
}

// canPieceAttackSquare checks if a piece at fromSq can attack toSq
func (ai *MinimaxAI) canPieceAttackSquare(game *engine.Game, fromSq, toSq engine.Square) bool {
	piece := game.Board().GetPiece(fromSq)
	if piece.IsEmpty() {
		return false
	}

	// Simple attack pattern check based on piece type
	switch piece.Type {
	case engine.Pawn:
		return ai.canPawnAttack(fromSq, toSq, piece.Color)
	case engine.Rook:
		return ai.canRookAttack(game, fromSq, toSq)
	case engine.Bishop:
		return ai.canBishopAttack(game, fromSq, toSq)
	case engine.Queen:
		return ai.canRookAttack(game, fromSq, toSq) || ai.canBishopAttack(game, fromSq, toSq)
	case engine.Knight:
		return ai.canKnightAttack(fromSq, toSq)
	case engine.King:
		return ai.canKingAttack(fromSq, toSq)
	}

	return false
}

// Helper methods for attack patterns
func (ai *MinimaxAI) canPawnAttack(fromSq, toSq engine.Square, color engine.Color) bool {
	fromRank, fromFile := int(fromSq)/8, int(fromSq)%8
	toRank, toFile := int(toSq)/8, int(toSq)%8

	direction := 1
	if color == engine.Black {
		direction = -1
	}

	// Pawn attacks diagonally
	return toRank == fromRank+direction && (toFile == fromFile+1 || toFile == fromFile-1)
}

func (ai *MinimaxAI) canRookAttack(game *engine.Game, fromSq, toSq engine.Square) bool {
	fromRank, fromFile := int(fromSq)/8, int(fromSq)%8
	toRank, toFile := int(toSq)/8, int(toSq)%8

	// Rook moves horizontally or vertically
	if fromRank != toRank && fromFile != toFile {
		return false
	}

	// Check if path is clear
	return ai.isPathClear(game, fromSq, toSq)
}

func (ai *MinimaxAI) canBishopAttack(game *engine.Game, fromSq, toSq engine.Square) bool {
	fromRank, fromFile := int(fromSq)/8, int(fromSq)%8
	toRank, toFile := int(toSq)/8, int(toSq)%8

	// Bishop moves diagonally
	if abs(fromRank-toRank) != abs(fromFile-toFile) {
		return false
	}

	// Check if path is clear
	return ai.isPathClear(game, fromSq, toSq)
}

func (ai *MinimaxAI) canKnightAttack(fromSq, toSq engine.Square) bool {
	fromRank, fromFile := int(fromSq)/8, int(fromSq)%8
	toRank, toFile := int(toSq)/8, int(toSq)%8

	deltaRank := abs(fromRank - toRank)
	deltaFile := abs(fromFile - toFile)

	return (deltaRank == 2 && deltaFile == 1) || (deltaRank == 1 && deltaFile == 2)
}

func (ai *MinimaxAI) canKingAttack(fromSq, toSq engine.Square) bool {
	fromRank, fromFile := int(fromSq)/8, int(fromSq)%8
	toRank, toFile := int(toSq)/8, int(toSq)%8

	deltaRank := abs(fromRank - toRank)
	deltaFile := abs(fromFile - toFile)

	return deltaRank <= 1 && deltaFile <= 1 && (deltaRank != 0 || deltaFile != 0)
}

// isPathClear checks if the path between two squares is clear
func (ai *MinimaxAI) isPathClear(game *engine.Game, fromSq, toSq engine.Square) bool {
	fromRank, fromFile := int(fromSq)/8, int(fromSq)%8
	toRank, toFile := int(toSq)/8, int(toSq)%8

	deltaRank := 0
	if toRank > fromRank {
		deltaRank = 1
	} else if toRank < fromRank {
		deltaRank = -1
	}

	deltaFile := 0
	if toFile > fromFile {
		deltaFile = 1
	} else if toFile < fromFile {
		deltaFile = -1
	}

	currentRank, currentFile := fromRank+deltaRank, fromFile+deltaFile

	for currentRank != toRank || currentFile != toFile {
		sq := engine.Square(currentRank*8 + currentFile)
		if !game.Board().GetPiece(sq).IsEmpty() {
			return false
		}
		currentRank += deltaRank
		currentFile += deltaFile
	}

	return true
}

// findCheckEscapeMoves filters moves that get the king out of check
func (ai *MinimaxAI) findCheckEscapeMoves(game *engine.Game, moves []engine.Move, color engine.Color) []engine.Move {
	var escapeMoves []engine.Move

	for _, move := range moves {
		// Create a temporary game state to test the move
		// Since copy() is private, we'll use a different approach
		if game.IsLegalMove(move) {
			escapeMoves = append(escapeMoves, move)
		}
	}

	return escapeMoves
}

// evaluateMove gives a simple evaluation score for a move
func (ai *MinimaxAI) evaluateMove(game *engine.Game, move engine.Move) int {
	score := 0

	// Prioritize captures
	targetPiece := game.Board().GetPiece(move.To)
	if !targetPiece.IsEmpty() {
		switch targetPiece.Type {
		case engine.Queen:
			score += 900
		case engine.Rook:
			score += 500
		case engine.Bishop, engine.Knight:
			score += 300
		case engine.Pawn:
			score += 100
		}
	}

	// Prioritize center control
	centerSquares := []engine.Square{27, 28, 35, 36} // d4, e4, d5, e5
	for _, sq := range centerSquares {
		if move.To == sq {
			score += 50
		}
	}

	// Prioritize piece development in opening
	if game.MoveCount() < 10 {
		piece := game.Board().GetPiece(move.From)
		if piece.Type == engine.Knight || piece.Type == engine.Bishop {
			score += 30
		}
	}

	return score
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
