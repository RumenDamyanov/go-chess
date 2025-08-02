package engine

import (
	"errors"
	"fmt"
	"strings"
)

// MoveType represents the type of chess move.
type MoveType int

const (
	// Normal represents a standard move.
	Normal MoveType = iota
	// Capture represents a capturing move.
	Capture
	// Castling represents a castling move.
	Castling
	// EnPassant represents an en passant capture.
	EnPassant
	// Promotion represents a pawn promotion.
	Promotion
)

// String returns the string representation of a move type.
func (mt MoveType) String() string {
	switch mt {
	case Normal:
		return "normal"
	case Capture:
		return "capture"
	case Castling:
		return "castling"
	case EnPassant:
		return "en_passant"
	case Promotion:
		return "promotion"
	default:
		return "unknown"
	}
}

// Move represents a chess move.
type Move struct {
	From      Square
	To        Square
	Type      MoveType
	Piece     Piece
	Captured  Piece
	Promotion PieceType
}

// String returns the string representation of a move in algebraic notation.
func (m Move) String() string {
	if m.Type == Castling {
		if m.To.File() > m.From.File() {
			return "O-O" // Kingside castling
		}
		return "O-O-O" // Queenside castling
	}

	notation := m.From.String() + m.To.String()

	if m.Type == Promotion {
		switch m.Promotion {
		case Queen:
			notation += "Q"
		case Rook:
			notation += "R"
		case Bishop:
			notation += "B"
		case Knight:
			notation += "N"
		}
	}

	return notation
}

// GameStatus represents the current status of the game.
type GameStatus int

const (
	// InProgress indicates the game is still in progress.
	InProgress GameStatus = iota
	// WhiteWins indicates white has won the game.
	WhiteWins
	// BlackWins indicates black has won the game.
	BlackWins
	// Draw indicates the game is a draw.
	Draw
)

// String returns the string representation of the game status.
func (gs GameStatus) String() string {
	switch gs {
	case InProgress:
		return "in_progress"
	case WhiteWins:
		return "white_wins"
	case BlackWins:
		return "black_wins"
	case Draw:
		return "draw"
	default:
		return "unknown"
	}
}

// CastlingRights represents the castling rights for both players.
type CastlingRights struct {
	WhiteKingside  bool
	WhiteQueenside bool
	BlackKingside  bool
	BlackQueenside bool
}

// Game represents a chess game state.
type Game struct {
	board           *Board
	activeColor     Color
	castlingRights  CastlingRights
	enPassantSquare Square
	halfMoveClock   int
	moveCount       int
	moveHistory     []Move
	status          GameStatus
}

// NewGame creates a new chess game with the standard starting position.
func NewGame() *Game {
	return &Game{
		board:       NewBoard(),
		activeColor: White,
		castlingRights: CastlingRights{
			WhiteKingside:  true,
			WhiteQueenside: true,
			BlackKingside:  true,
			BlackQueenside: true,
		},
		enPassantSquare: -1,
		halfMoveClock:   0,
		moveCount:       1,
		moveHistory:     make([]Move, 0),
		status:          InProgress,
	}
}

// Board returns a copy of the current board.
func (g *Game) Board() *Board {
	return g.board.Copy()
}

// ActiveColor returns the color of the player whose turn it is.
func (g *Game) ActiveColor() Color {
	return g.activeColor
}

// Status returns the current game status.
func (g *Game) Status() GameStatus {
	return g.status
}

// MoveCount returns the current move count.
func (g *Game) MoveCount() int {
	return g.moveCount
}

// MoveHistory returns a copy of the move history.
func (g *Game) MoveHistory() []Move {
	history := make([]Move, len(g.moveHistory))
	copy(history, g.moveHistory)
	return history
}

// ParseMove parses a move from algebraic notation (e.g., "e2e4", "e7e8Q").
func (g *Game) ParseMove(notation string) (Move, error) {
	notation = strings.TrimSpace(notation)

	// Handle castling notation
	if notation == "O-O" || notation == "0-0" {
		return g.parseCastlingMove(true)
	}
	if notation == "O-O-O" || notation == "0-0-0" {
		return g.parseCastlingMove(false)
	}

	// Standard notation: e2e4, e7e8Q
	if len(notation) < 4 {
		return Move{}, errors.New("invalid move notation")
	}

	fromStr := notation[:2]
	toStr := notation[2:4]

	from, err := SquareFromString(fromStr)
	if err != nil {
		return Move{}, fmt.Errorf("invalid from square: %w", err)
	}

	to, err := SquareFromString(toStr)
	if err != nil {
		return Move{}, fmt.Errorf("invalid to square: %w", err)
	}

	piece := g.board.GetPiece(from)
	if piece.IsEmpty() || piece.Color != g.activeColor {
		return Move{}, errors.New("no piece to move or wrong color")
	}

	captured := g.board.GetPiece(to)
	moveType := Normal
	if !captured.IsEmpty() {
		moveType = Capture
	}

	move := Move{
		From:     from,
		To:       to,
		Type:     moveType,
		Piece:    piece,
		Captured: captured,
	}

	// Check for promotion
	if len(notation) == 5 && piece.Type == Pawn {
		promChar := strings.ToUpper(notation[4:5])
		switch promChar {
		case "Q":
			move.Promotion = Queen
		case "R":
			move.Promotion = Rook
		case "B":
			move.Promotion = Bishop
		case "N":
			move.Promotion = Knight
		default:
			return Move{}, errors.New("invalid promotion piece")
		}
		move.Type = Promotion
	}

	return move, nil
}

// parseCastlingMove parses a castling move.
func (g *Game) parseCastlingMove(kingside bool) (Move, error) {
	var kingFrom, kingTo Square

	if g.activeColor == White {
		kingFrom = E1
		if kingside {
			kingTo = G1
		} else {
			kingTo = C1
		}
	} else {
		kingFrom = E8
		if kingside {
			kingTo = G8
		} else {
			kingTo = C8
		}
	}

	king := g.board.GetPiece(kingFrom)
	if king.Type != King || king.Color != g.activeColor {
		return Move{}, errors.New("king not in position for castling")
	}

	return Move{
		From:  kingFrom,
		To:    kingTo,
		Type:  Castling,
		Piece: king,
	}, nil
}

// IsLegalMove checks if a move is legal in the current position.
func (g *Game) IsLegalMove(move Move) bool {
	// Basic validation
	piece := g.board.GetPiece(move.From)
	if piece.IsEmpty() || piece.Color != g.activeColor {
		return false
	}

	// Check if the move is pseudo-legal for the piece type
	if !g.isPseudoLegalMove(move) {
		return false
	}

	// Make a copy of the game to test the move
	gameCopy := g.copy()
	gameCopy.makeMove(move)

	// Check if the king is in check after the move
	return !gameCopy.isInCheck(g.activeColor)
}

// MakeMove makes a move if it's legal.
func (g *Game) MakeMove(move Move) error {
	if !g.IsLegalMove(move) {
		return errors.New("illegal move")
	}

	g.makeMove(move)
	g.moveHistory = append(g.moveHistory, move)

	// Switch active color
	if g.activeColor == White {
		g.activeColor = Black
	} else {
		g.activeColor = White
		g.moveCount++
	}

	// Update game status
	g.updateGameStatus()

	return nil
}

// makeMove executes a move without validation.
func (g *Game) makeMove(move Move) {
	// Handle castling
	if move.Type == Castling {
		g.executeCastling(move)
		return
	}

	// Handle en passant
	if move.Type == EnPassant {
		g.executeEnPassant(move)
		return
	}

	// Regular move
	g.board.SetPiece(move.To, move.Piece)
	g.board.SetPiece(move.From, Piece{Type: Empty})

	// Handle promotion
	if move.Type == Promotion {
		g.board.SetPiece(move.To, Piece{Type: move.Promotion, Color: move.Piece.Color})
	}

	// Update castling rights
	g.updateCastlingRights(move)

	// Update en passant square
	g.updateEnPassantSquare(move)

	// Update half-move clock
	g.updateHalfMoveClock(move)
}

// executeCastling executes a castling move.
func (g *Game) executeCastling(move Move) {
	// Move the king
	g.board.SetPiece(move.To, move.Piece)
	g.board.SetPiece(move.From, Piece{Type: Empty})

	// Move the rook
	var rookFrom, rookTo Square
	if move.To.File() > move.From.File() { // Kingside
		if g.activeColor == White {
			rookFrom, rookTo = H1, F1
		} else {
			rookFrom, rookTo = H8, F8
		}
	} else { // Queenside
		if g.activeColor == White {
			rookFrom, rookTo = A1, D1
		} else {
			rookFrom, rookTo = A8, D8
		}
	}

	rook := g.board.GetPiece(rookFrom)
	g.board.SetPiece(rookTo, rook)
	g.board.SetPiece(rookFrom, Piece{Type: Empty})
}

// executeEnPassant executes an en passant capture.
func (g *Game) executeEnPassant(move Move) {
	// Move the pawn
	g.board.SetPiece(move.To, move.Piece)
	g.board.SetPiece(move.From, Piece{Type: Empty})

	// Remove the captured pawn
	var capturedSquare Square
	if g.activeColor == White {
		capturedSquare = Square(int(move.To) - 8)
	} else {
		capturedSquare = Square(int(move.To) + 8)
	}
	g.board.SetPiece(capturedSquare, Piece{Type: Empty})
}

// isPseudoLegalMove checks if a move is pseudo-legal (doesn't check for check).
func (g *Game) isPseudoLegalMove(move Move) bool {
	piece := move.Piece

	switch piece.Type {
	case Pawn:
		return g.isPawnMoveLegal(move)
	case Rook:
		return g.isRookMoveLegal(move)
	case Knight:
		return g.isKnightMoveLegal(move)
	case Bishop:
		return g.isBishopMoveLegal(move)
	case Queen:
		return g.isQueenMoveLegal(move)
	case King:
		return g.isKingMoveLegal(move)
	}

	return false
}

// Helper methods for piece-specific move validation would go here...
// For brevity, I'll implement basic versions:

func (g *Game) isPawnMoveLegal(move Move) bool {
	// Simplified pawn move validation
	direction := 1
	if move.Piece.Color == Black {
		direction = -1
	}

	fromRank := move.From.Rank()
	toRank := move.To.Rank()
	fileDiff := abs(move.To.File() - move.From.File())

	// Forward move
	if fileDiff == 0 {
		if toRank-fromRank == direction && g.board.GetPiece(move.To).IsEmpty() {
			return true
		}
		// Two squares from starting position
		if (fromRank == 1 && move.Piece.Color == White) || (fromRank == 6 && move.Piece.Color == Black) {
			if toRank-fromRank == 2*direction && g.board.GetPiece(move.To).IsEmpty() {
				return true
			}
		}
	}

	// Diagonal capture
	if fileDiff == 1 && toRank-fromRank == direction {
		target := g.board.GetPiece(move.To)
		if !target.IsEmpty() && target.Color != move.Piece.Color {
			return true
		}
		// En passant capture
		if move.Type == EnPassant && move.To == g.enPassantSquare {
			return true
		}
	}

	return false
}

func (g *Game) isRookMoveLegal(move Move) bool {
	return g.isPathClear(move.From, move.To) &&
		(move.From.Rank() == move.To.Rank() || move.From.File() == move.To.File())
}

func (g *Game) isKnightMoveLegal(move Move) bool {
	fileDiff := abs(move.To.File() - move.From.File())
	rankDiff := abs(move.To.Rank() - move.From.Rank())
	return (fileDiff == 2 && rankDiff == 1) || (fileDiff == 1 && rankDiff == 2)
}

func (g *Game) isBishopMoveLegal(move Move) bool {
	fileDiff := abs(move.To.File() - move.From.File())
	rankDiff := abs(move.To.Rank() - move.From.Rank())
	return fileDiff == rankDiff && g.isPathClear(move.From, move.To)
}

func (g *Game) isQueenMoveLegal(move Move) bool {
	return g.isRookMoveLegal(move) || g.isBishopMoveLegal(move)
}

func (g *Game) isKingMoveLegal(move Move) bool {
	if move.Type == Castling {
		return g.canCastle(move.To.File() > move.From.File())
	}

	fileDiff := abs(move.To.File() - move.From.File())
	rankDiff := abs(move.To.Rank() - move.From.Rank())

	// King can only move one square in any direction
	if fileDiff > 1 || rankDiff > 1 {
		return false
	}

	// Check if destination is empty or contains opponent piece
	target := g.board.GetPiece(move.To)
	if !target.IsEmpty() && target.Color == move.Piece.Color {
		return false
	}

	return true
}

// isPathClear checks if the path between two squares is clear.
func (g *Game) isPathClear(from, to Square) bool {
	fileDiff := to.File() - from.File()
	rankDiff := to.Rank() - from.Rank()

	fileStep := sign(fileDiff)
	rankStep := sign(rankDiff)

	current := Square(int(from) + fileStep + rankStep*8)

	for current != to {
		if !g.board.GetPiece(current).IsEmpty() {
			return false
		}
		current = Square(int(current) + fileStep + rankStep*8)
	}

	return true
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// Placeholder implementations for complex methods
func (g *Game) updateCastlingRights(move Move) {
	// Implementation would update castling rights based on the move
}

func (g *Game) updateEnPassantSquare(move Move) {
	g.enPassantSquare = -1 // Reset en passant square

	// Set en passant square for pawn double moves
	if move.Piece.Type == Pawn && abs(move.To.Rank()-move.From.Rank()) == 2 {
		g.enPassantSquare = Square((int(move.From) + int(move.To)) / 2)
	}
}

func (g *Game) updateHalfMoveClock(move Move) {
	if move.Piece.Type == Pawn || move.Type == Capture {
		g.halfMoveClock = 0
	} else {
		g.halfMoveClock++
	}
}

func (g *Game) canCastle(kingside bool) bool {
	// Simplified castling check
	if g.activeColor == White {
		if kingside {
			return g.castlingRights.WhiteKingside
		}
		return g.castlingRights.WhiteQueenside
	}

	if kingside {
		return g.castlingRights.BlackKingside
	}
	return g.castlingRights.BlackQueenside
}

func (g *Game) isInCheck(color Color) bool {
	// Find the king
	kingSquare := Square(-1)
	for sq := Square(0); sq < 64; sq++ {
		piece := g.board.GetPiece(sq)
		if piece.Type == King && piece.Color == color {
			kingSquare = sq
			break
		}
	}

	if kingSquare == -1 {
		return false // King not found
	}

	// Check if any opponent piece can attack the king
	// This is a simplified implementation
	return false
}

func (g *Game) updateGameStatus() {
	// Check for checkmate, stalemate, draw conditions
	// This is a placeholder implementation
	g.status = InProgress
}

func (g *Game) copy() *Game {
	newGame := &Game{
		board:           g.board.Copy(),
		activeColor:     g.activeColor,
		castlingRights:  g.castlingRights,
		enPassantSquare: g.enPassantSquare,
		halfMoveClock:   g.halfMoveClock,
		moveCount:       g.moveCount,
		status:          g.status,
	}

	newGame.moveHistory = make([]Move, len(g.moveHistory))
	copy(newGame.moveHistory, g.moveHistory)

	return newGame
}
