package engine

import (
	"errors"
	"fmt"
	"strconv"
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
	// Check indicates the current player's king is in check.
	Check
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
	case Check:
		return "check"
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
	// startedFromFEN indicates the game began (or was reset) from a custom FEN
	startedFromFEN bool
	// startingFEN stores the original FEN the current game was loaded from (if any)
	startingFEN string
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
		startedFromFEN:  false,
		startingFEN:     "",
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
	} else if piece.Type == Pawn && from.File() != to.File() { // diagonal pawn move to empty square => possible en passant
		// Valid en passant if target square equals enPassantSquare
		if g.enPassantSquare != -1 && to == g.enPassantSquare {
			moveType = EnPassant
			// Captured pawn is behind the target square
			var capSq Square
			if piece.Color == White {
				capSq = to - 8
			} else {
				capSq = to + 8
			}
			captured = g.board.GetPiece(capSq)
		}
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

	// Check if the destination contains a king - capturing the king should never be allowed
	targetPiece := g.board.GetPiece(move.To)
	if !targetPiece.IsEmpty() && targetPiece.Type == King {
		return false // Cannot capture the king
	}

	// Check if the move is pseudo-legal for the piece type
	if !g.isPseudoLegalMove(move) {
		return false
	}

	// Make a copy of the game to test the move
	gameCopy := g.copy()
	gameCopy.makeMoveWithoutStatusUpdate(move)

	// Check if the king is in check after the move
	inCheck := gameCopy.isInCheck(g.activeColor)

	return !inCheck
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

// makeMoveWithoutStatusUpdate executes a move without validation or status update.
// This is used internally for move validation to avoid infinite recursion.
func (g *Game) makeMoveWithoutStatusUpdate(move Move) {
	// Handle castling
	if move.Type == Castling {
		g.executeCastling(move)
		// Remove castling rights for the moving side (king moved)
		g.updateCastlingRights(move)
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

	// Switch active color for the copy
	if g.activeColor == White {
		g.activeColor = Black
	} else {
		g.activeColor = White
	}
}

// makeMove executes a move without validation.
func (g *Game) makeMove(move Move) {
	// Handle castling
	if move.Type == Castling {
		g.executeCastling(move)
		// Update castling rights after executing castling (king moved)
		g.updateCastlingRights(move)
		// No en passant square or half-move clock change beyond standard; handle them similarly to normal moves
		g.updateEnPassantSquare(move)
		g.updateHalfMoveClock(move)
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
		result := g.isPawnMoveLegal(move)
		return result
	case Rook:
		result := g.isRookMoveLegal(move)
		return result
	case Knight:
		result := g.isKnightMoveLegal(move)
		return result
	case Bishop:
		result := g.isBishopMoveLegal(move)
		return result
	case Queen:
		result := g.isQueenMoveLegal(move)
		return result
	case King:
		result := g.isKingMoveLegal(move)
		return result
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
	// If a king moves, remove both rights for that color
	if move.Piece.Type == King {
		if move.Piece.Color == White {
			g.castlingRights.WhiteKingside = false
			g.castlingRights.WhiteQueenside = false
		} else {
			g.castlingRights.BlackKingside = false
			g.castlingRights.BlackQueenside = false
		}
	}

	// If a rook moves from its original square, remove that side's right
	if move.Piece.Type == Rook {
		switch move.From {
		case H1:
			g.castlingRights.WhiteKingside = false
		case A1:
			g.castlingRights.WhiteQueenside = false
		case H8:
			g.castlingRights.BlackKingside = false
		case A8:
			g.castlingRights.BlackQueenside = false
		}
	}

	// If a rook is captured on its original square, remove that side's right
	if move.Type == Capture && move.Captured.Type == Rook {
		switch move.To { // capture destination square contains captured rook
		case H1:
			g.castlingRights.WhiteKingside = false
		case A1:
			g.castlingRights.WhiteQueenside = false
		case H8:
			g.castlingRights.BlackKingside = false
		case A8:
			g.castlingRights.BlackQueenside = false
		}
	}

	// If castling move, move the rook implicitly handled elsewhere; rights removed via king move logic above
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
	// Use the detailed castling validation functions
	if kingside {
		return g.canCastleKingside(g.activeColor)
	}
	return g.canCastleQueenside(g.activeColor)
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
	opponentColor := White
	if color == White {
		opponentColor = Black
	}

	// Check all opponent pieces to see if they can attack the king
	for sq := Square(0); sq < 64; sq++ {
		piece := g.board.GetPiece(sq)
		if piece.IsEmpty() || piece.Color != opponentColor {
			continue
		}

		// Generate pseudo-legal moves for the opponent piece
		moves := g.generatePseudoLegalMoves(sq, piece)

		// Check if any move attacks the king
		for _, move := range moves {
			if move.To == kingSquare {
				return true
			}
		}
	}

	return false
}

// GetAllLegalMoves generates all legal moves for the current player
func (g *Game) GetAllLegalMoves() []Move {
	var legalMoves []Move

	// Iterate through all squares
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			square := Square(rank*8 + file)
			piece := g.board.GetPiece(square)

			// Skip empty squares and opponent pieces
			if piece.IsEmpty() || piece.Color != g.activeColor {
				continue
			}

			// Generate pseudo-legal moves for this piece
			moves := g.generatePseudoLegalMoves(square, piece)

			// Filter out illegal moves (those that leave king in check)
			for _, move := range moves {
				if g.IsLegalMove(move) {
					legalMoves = append(legalMoves, move)
				}
			}
		}
	}

	return legalMoves
}

// generatePseudoLegalMoves generates all pseudo-legal moves for a piece at the given square
func (g *Game) generatePseudoLegalMoves(from Square, piece Piece) []Move {
	var moves []Move
	pieceType := piece.Type

	switch pieceType {
	case Pawn:
		moves = append(moves, g.generatePawnMoves(from)...)
	case Rook:
		moves = append(moves, g.generateRookMoves(from)...)
	case Knight:
		moves = append(moves, g.generateKnightMoves(from)...)
	case Bishop:
		moves = append(moves, g.generateBishopMoves(from)...)
	case Queen:
		moves = append(moves, g.generateQueenMoves(from)...)
	case King:
		moves = append(moves, g.generateKingMoves(from)...)
	}

	return moves
}

// generatePawnMoves generates all pseudo-legal pawn moves
func (g *Game) generatePawnMoves(from Square) []Move {
	var moves []Move
	piece := g.board.GetPiece(from)
	color := piece.Color

	direction := 1
	startRank := 1
	if color == Black {
		direction = -1
		startRank = 6
	}

	rank := int(from / 8)
	file := int(from % 8)

	// Forward move
	toSquare := Square((rank+direction)*8 + file)
	if rank+direction >= 0 && rank+direction < 8 && g.board.GetPiece(toSquare).IsEmpty() {
		moves = append(moves, Move{From: from, To: toSquare, Type: Normal, Piece: piece})

		// Double move from starting position
		if rank == startRank {
			toSquare2 := Square((rank+2*direction)*8 + file)
			if g.board.GetPiece(toSquare2).IsEmpty() {
				moves = append(moves, Move{From: from, To: toSquare2, Type: Normal, Piece: piece})
			}
		}
	}

	// Captures
	for _, fileOffset := range []int{-1, 1} {
		newFile := file + fileOffset
		newRank := rank + direction
		if newFile >= 0 && newFile < 8 && newRank >= 0 && newRank < 8 {
			toSquare := Square(newRank*8 + newFile)
			targetPiece := g.board.GetPiece(toSquare)
			if !targetPiece.IsEmpty() && targetPiece.Color != color {
				moves = append(moves, Move{From: from, To: toSquare, Type: Normal, Piece: piece})
			}
		}
	}

	return moves
}

// generateSlidingMoves generates moves for sliding pieces (rook, bishop, queen)
func (g *Game) generateSlidingMoves(from Square, directions [][]int) []Move {
	var moves []Move
	piece := g.board.GetPiece(from)
	color := piece.Color

	rank := int(from / 8)
	file := int(from % 8)

	for _, dir := range directions {
		for i := 1; i < 8; i++ {
			newRank := rank + dir[0]*i
			newFile := file + dir[1]*i

			if newRank < 0 || newRank >= 8 || newFile < 0 || newFile >= 8 {
				break
			}

			toSquare := Square(newRank*8 + newFile)
			targetPiece := g.board.GetPiece(toSquare)

			if targetPiece.IsEmpty() {
				moves = append(moves, Move{From: from, To: toSquare, Type: Normal, Piece: piece})
			} else {
				if targetPiece.Color != color {
					moves = append(moves, Move{From: from, To: toSquare, Type: Normal, Piece: piece})
				}
				break
			}
		}
	}

	return moves
}

// generateKnightMoves generates all pseudo-legal knight moves
func (g *Game) generateKnightMoves(from Square) []Move {
	var moves []Move
	piece := g.board.GetPiece(from)
	color := piece.Color

	rank := int(from / 8)
	file := int(from % 8)

	knightMoves := [][]int{{2, 1}, {2, -1}, {-2, 1}, {-2, -1}, {1, 2}, {1, -2}, {-1, 2}, {-1, -2}}

	for _, move := range knightMoves {
		newRank := rank + move[0]
		newFile := file + move[1]

		if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
			toSquare := Square(newRank*8 + newFile)
			targetPiece := g.board.GetPiece(toSquare)

			if targetPiece.IsEmpty() || targetPiece.Color != color {
				moves = append(moves, Move{From: from, To: toSquare, Type: Normal, Piece: piece})
			}
		}
	}

	return moves
}

// generateRookMoves generates all pseudo-legal rook moves
func (g *Game) generateRookMoves(from Square) []Move {
	return g.generateSlidingMoves(from, [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}})
}

// generateBishopMoves generates all pseudo-legal bishop moves
func (g *Game) generateBishopMoves(from Square) []Move {
	return g.generateSlidingMoves(from, [][]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
}

// generateQueenMoves generates all pseudo-legal queen moves
func (g *Game) generateQueenMoves(from Square) []Move {
	return g.generateSlidingMoves(from, [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
}

// generateKingMoves generates all pseudo-legal king moves
func (g *Game) generateKingMoves(from Square) []Move {
	var moves []Move
	piece := g.board.GetPiece(from)
	color := piece.Color

	rank := int(from / 8)
	file := int(from % 8)

	kingMoves := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	for _, move := range kingMoves {
		newRank := rank + move[0]
		newFile := file + move[1]

		if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
			toSquare := Square(newRank*8 + newFile)
			targetPiece := g.board.GetPiece(toSquare)

			if targetPiece.IsEmpty() || targetPiece.Color != color {
				moves = append(moves, Move{From: from, To: toSquare, Type: Normal, Piece: piece})
			}
		}
	}

	// Add castling moves
	moves = append(moves, g.generateCastlingMoves(from)...)

	return moves
}

// generateCastlingMoves generates castling moves for the king
func (g *Game) generateCastlingMoves(from Square) []Move {
	var moves []Move
	piece := g.board.GetPiece(from)
	color := piece.Color

	// Only generate castling moves if the king is on its starting square
	expectedKingSquare := E1
	if color == Black {
		expectedKingSquare = E8
	}

	if from != expectedKingSquare {
		return moves
	}

	// Try kingside castling
	if g.canCastleKingside(color) {
		kingsideMove, err := g.parseCastlingMove(true)
		if err == nil {
			moves = append(moves, kingsideMove)
		}
	}

	// Try queenside castling
	if g.canCastleQueenside(color) {
		queensideMove, err := g.parseCastlingMove(false)
		if err == nil {
			moves = append(moves, queensideMove)
		}
	}

	return moves
}

// canCastleKingside checks if kingside castling is possible for the given color
func (g *Game) canCastleKingside(color Color) bool {
	// Check castling rights
	if color == White && !g.castlingRights.WhiteKingside {
		return false
	}
	if color == Black && !g.castlingRights.BlackKingside {
		return false
	}

	// Check if squares between king and rook are empty
	kingSquare := E1
	rookSquare := H1
	if color == Black {
		kingSquare = E8
		rookSquare = H8
	}

	// Check squares F and G (between king and rook)
	for square := kingSquare + 1; square < rookSquare; square++ {
		if !g.board.GetPiece(square).IsEmpty() {
			return false
		}
	}

	// Check if king is currently in check
	if g.isInCheck(color) {
		return false
	}

	// Check if king passes through or ends up in check
	for square := kingSquare; square <= kingSquare+2; square++ {
		// Create a temporary move to test
		tempMove := Move{From: kingSquare, To: square, Type: Normal, Piece: g.board.GetPiece(kingSquare)}
		if square != kingSquare && g.wouldBeInCheckAfterMove(tempMove, color) {
			return false
		}
	}

	return true
}

// canCastleQueenside checks if queenside castling is possible for the given color
func (g *Game) canCastleQueenside(color Color) bool {
	// Check castling rights
	if color == White && !g.castlingRights.WhiteQueenside {
		return false
	}
	if color == Black && !g.castlingRights.BlackQueenside {
		return false
	}

	// Check if squares between king and rook are empty
	kingSquare := E1
	rookSquare := A1
	if color == Black {
		kingSquare = E8
		rookSquare = A8
	}

	// Check squares B, C, D (between rook and king)
	for square := rookSquare + 1; square < kingSquare; square++ {
		if !g.board.GetPiece(square).IsEmpty() {
			return false
		}
	}

	// Check if king is currently in check
	if g.isInCheck(color) {
		return false
	}

	// Check if king passes through or ends up in check
	for square := kingSquare; square >= kingSquare-2; square-- {
		// Create a temporary move to test
		tempMove := Move{From: kingSquare, To: square, Type: Normal, Piece: g.board.GetPiece(kingSquare)}
		if square != kingSquare && g.wouldBeInCheckAfterMove(tempMove, color) {
			return false
		}
	}

	return true
}

// wouldBeInCheckAfterMove checks if the king would be in check after a given move
func (g *Game) wouldBeInCheckAfterMove(move Move, kingColor Color) bool {
	// Create a copy of the game to test the move
	gameCopy := g.copy()

	// Make the move on the copy using makeMoveWithoutStatusUpdate to avoid recursion
	gameCopy.makeMoveWithoutStatusUpdate(move)

	// Check if the king is in check in the resulting position
	return gameCopy.isInCheck(kingColor)
}

func (g *Game) updateGameStatus() {
	// Check for checkmate, stalemate, draw conditions
	legalMoves := g.GetAllLegalMoves()

	if len(legalMoves) == 0 {
		// No legal moves available
		if g.isInCheck(g.activeColor) {
			// King is in check and has no legal moves = checkmate
			if g.activeColor == White {
				g.status = BlackWins
			} else {
				g.status = WhiteWins
			}
		} else {
			// King is not in check but has no legal moves = stalemate
			g.status = Draw
		}
	} else {
		// Game continues - check if king is in check
		if g.isInCheck(g.activeColor) {
			g.status = Check
		} else {
			g.status = InProgress
		}
	}
}

// ToFEN converts the current game position to FEN (Forsyth-Edwards Notation).
func (g *Game) ToFEN() string {
	var fen strings.Builder

	// 1. Piece placement
	for rank := 7; rank >= 0; rank-- {
		emptyCount := 0
		for file := 0; file < 8; file++ {
			square := Square(rank*8 + file)
			piece := g.board.GetPiece(square)

			if piece.IsEmpty() {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(fmt.Sprintf("%d", emptyCount))
					emptyCount = 0
				}
				fen.WriteString(g.pieceToFENChar(piece))
			}
		}
		if emptyCount > 0 {
			fen.WriteString(fmt.Sprintf("%d", emptyCount))
		}
		if rank > 0 {
			fen.WriteString("/")
		}
	}

	// 2. Active color
	fen.WriteString(" ")
	if g.activeColor == White {
		fen.WriteString("w")
	} else {
		fen.WriteString("b")
	}

	// 3. Castling rights
	fen.WriteString(" ")
	castling := ""
	if g.castlingRights.WhiteKingside {
		castling += "K"
	}
	if g.castlingRights.WhiteQueenside {
		castling += "Q"
	}
	if g.castlingRights.BlackKingside {
		castling += "k"
	}
	if g.castlingRights.BlackQueenside {
		castling += "q"
	}
	if castling == "" {
		castling = "-"
	}
	fen.WriteString(castling)

	// 4. En passant square
	fen.WriteString(" ")
	if g.enPassantSquare == -1 {
		fen.WriteString("-")
	} else {
		fen.WriteString(g.enPassantSquare.String())
	}

	// 5. Half-move clock
	fen.WriteString(fmt.Sprintf(" %d", g.halfMoveClock))

	// 6. Full-move number
	fen.WriteString(fmt.Sprintf(" %d", g.moveCount))

	return fen.String()
}

// pieceToFENChar converts a piece to its FEN character representation.
func (g *Game) pieceToFENChar(piece Piece) string {
	var char string
	switch piece.Type {
	case Pawn:
		char = "p"
	case Rook:
		char = "r"
	case Knight:
		char = "n"
	case Bishop:
		char = "b"
	case Queen:
		char = "q"
	case King:
		char = "k"
	default:
		return ""
	}

	if piece.Color == White {
		return strings.ToUpper(char)
	}
	return char
}

// ParseFEN loads a position from a FEN string into the current game.
// Supported fields: piece placement, active color, castling rights, en passant square,
// halfmove clock, fullmove number. Move history and status are reset and then status recalculated.
func (g *Game) ParseFEN(fen string) error {
	parts := strings.Fields(strings.TrimSpace(fen))
	if len(parts) < 4 {
		return fmt.Errorf("invalid FEN: expected at least 4 fields, got %d", len(parts))
	}

	// 1. Piece placement
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return fmt.Errorf("invalid FEN: expected 8 ranks, got %d", len(ranks))
	}

	// Clear board first
	for i := 0; i < 64; i++ {
		g.board.squares[i] = Piece{Type: Empty}
	}

	for rankIdx, rankStr := range ranks { // rankIdx 0 = rank 8 in FEN
		file := 0
		for _, ch := range rankStr {
			if file > 7 {
				return fmt.Errorf("invalid FEN rank %d: too many squares", 8-rankIdx)
			}
			if ch >= '1' && ch <= '8' {
				skip := int(ch - '0')
				file += skip
				continue
			}
			var pieceType PieceType
			var color Color
			switch ch {
			case 'p', 'P':
				pieceType = Pawn
			case 'r', 'R':
				pieceType = Rook
			case 'n', 'N':
				pieceType = Knight
			case 'b', 'B':
				pieceType = Bishop
			case 'q', 'Q':
				pieceType = Queen
			case 'k', 'K':
				pieceType = King
			default:
				return fmt.Errorf("invalid FEN piece character: %c", ch)
			}
			if ch >= 'A' && ch <= 'Z' {
				color = White
			} else {
				color = Black
			}
			square := Square((7-rankIdx)*8 + file)
			g.board.squares[square] = Piece{Type: pieceType, Color: color}
			file++
		}
		if file != 8 {
			return fmt.Errorf("invalid FEN rank %d: expected 8 files, got %d", 8-rankIdx, file)
		}
	}

	// 2. Active color
	if len(parts[1]) != 1 || (parts[1] != "w" && parts[1] != "b") {
		return fmt.Errorf("invalid FEN active color: %s", parts[1])
	}
	if parts[1] == "w" {
		g.activeColor = White
	} else {
		g.activeColor = Black
	}

	// 3. Castling rights
	castling := parts[2]
	g.castlingRights = CastlingRights{}
	if castling != "-" {
		for _, ch := range castling {
			switch ch {
			case 'K':
				g.castlingRights.WhiteKingside = true
			case 'Q':
				g.castlingRights.WhiteQueenside = true
			case 'k':
				g.castlingRights.BlackKingside = true
			case 'q':
				g.castlingRights.BlackQueenside = true
			default:
				return fmt.Errorf("invalid castling char: %c", ch)
			}
		}
	}

	// 4. En passant square
	enPassant := parts[3]
	if enPassant == "-" {
		g.enPassantSquare = -1
	} else {
		if len(enPassant) != 2 {
			return fmt.Errorf("invalid en-passant square: %s", enPassant)
		}
		sq, err := SquareFromString(enPassant)
		if err != nil {
			return fmt.Errorf("invalid en-passant square: %w", err)
		}
		g.enPassantSquare = sq
	}

	// 5 & 6 (optional) halfmove clock and fullmove number
	g.halfMoveClock = 0
	g.moveCount = 1
	if len(parts) >= 5 {
		hm, err := strconv.Atoi(parts[4])
		if err != nil || hm < 0 {
			return fmt.Errorf("invalid halfmove clock: %s", parts[4])
		}
		g.halfMoveClock = hm
	}
	if len(parts) >= 6 {
		fm, err := strconv.Atoi(parts[5])
		if err != nil || fm < 1 {
			return fmt.Errorf("invalid fullmove number: %s", parts[5])
		}
		g.moveCount = fm
	}

	// Reset move history and recalc status
	g.moveHistory = nil
	g.status = InProgress
	g.startedFromFEN = true
	g.startingFEN = fen
	g.updateGameStatus()
	return nil
}

// StartedFromFEN returns true if the current game originated from a custom FEN.
func (g *Game) StartedFromFEN() bool { return g.startedFromFEN }

// StartingFEN returns the original starting FEN if provided.
func (g *Game) StartingFEN() string { return g.startingFEN }

// Evaluate returns a simple material + activity evaluation (centipawns from White's perspective).
func (g *Game) Evaluate() int {
	values := map[PieceType]int{
		Pawn:   100,
		Knight: 320,
		Bishop: 330,
		Rook:   500,
		Queen:  900,
		King:   0,
	}
	score := 0
	for sq := Square(0); sq < 64; sq++ {
		p := g.board.GetPiece(sq)
		if p.IsEmpty() {
			continue
		}
		v := values[p.Type]
		if p.Color == White {
			score += v
		} else {
			score -= v
		}
		// Small central control bonus
		file := sq.File()
		rank := sq.Rank()
		if file >= 2 && file <= 5 && rank >= 2 && rank <= 5 { // 16 central squares
			if p.Color == White {
				score += 5
			} else {
				score -= 5
			}
		}
	}
	return score
}

// GenerateSAN returns SAN strings for the game's move history.
// It reconstructs moves from the starting position (initial or loaded FEN) to ensure correctness.
func (g *Game) GenerateSAN() []string {
	san := make([]string, 0, len(g.moveHistory))
	// Recreate starting position
	var replay *Game
	if g.startedFromFEN && g.startingFEN != "" {
		replay = NewGame()
		_ = replay.ParseFEN(g.startingFEN) // ignore error: stored FEN assumed valid
	} else {
		replay = NewGame()
	}
	for _, mv := range g.moveHistory {
		san = append(san, replay.sanForMove(mv))
		// Apply move to advance position
		_ = replay.MakeMove(mv) // moves are assumed legal as they occurred in original game
	}
	return san
}

// sanForMove computes SAN for a move given the current position (before move is applied).
func (g *Game) sanForMove(m Move) string {
	piece := g.board.GetPiece(m.From)
	if piece.Type == King && m.Type == Castling {
		if m.To.File() > m.From.File() {
			return "O-O"
		}
		return "O-O-O"
	}

	// Determine if capture (include en passant)
	target := g.board.GetPiece(m.To)
	isCapture := (!target.IsEmpty() && target.Color != piece.Color) || m.Type == Capture || m.Type == EnPassant

	var sb strings.Builder
	if piece.Type == Pawn {
		// Pawn moves
		if isCapture {
			sb.WriteByte(byte('a' + m.From.File()))
			sb.WriteByte('x')
		}
		sb.WriteString(m.To.String())
		if m.Type == Promotion && m.Promotion != Empty {
			sb.WriteByte('=')
			// Map promotion piece to uppercase SAN letter
			switch m.Promotion {
			case Queen:
				sb.WriteString("Q")
			case Rook:
				sb.WriteString("R")
			case Bishop:
				sb.WriteString("B")
			case Knight:
				sb.WriteString("N")
			default:
				sb.WriteString("?")
			}
		}
	} else { // Piece moves (N,B,R,Q,K)
		// Map piece type to SAN letter
		switch piece.Type {
		case Knight:
			sb.WriteString("N")
		case Bishop:
			sb.WriteString("B")
		case Rook:
			sb.WriteString("R")
		case Queen:
			sb.WriteString("Q")
		case King:
			sb.WriteString("K")
		default:
			sb.WriteString("?")
		}
		// Need potential disambiguation
		needFile, needRank := g.disambiguation(piece, m)
		if needFile {
			sb.WriteByte(byte('a' + m.From.File()))
		}
		if needRank {
			sb.WriteByte(byte('1' + m.From.Rank()))
		}
		if isCapture {
			sb.WriteByte('x')
		}
		sb.WriteString(m.To.String())
		if m.Type == Promotion && m.Promotion != Empty {
			sb.WriteByte('=')
			switch m.Promotion {
			case Queen:
				sb.WriteString("Q")
			case Rook:
				sb.WriteString("R")
			case Bishop:
				sb.WriteString("B")
			case Knight:
				sb.WriteString("N")
			default:
				sb.WriteString("?")
			}
		}
	}

	// Determine check / mate after move using full MakeMove (which switches side & updates status)
	gameCopy := g.copy()
	_ = gameCopy.MakeMove(m) // ignore error; original move already known legal
	if gameCopy.status == WhiteWins || gameCopy.status == BlackWins {
		sb.WriteByte('#')
	} else if gameCopy.isInCheck(gameCopy.activeColor) { // after MakeMove, activeColor is opponent
		sb.WriteByte('+')
	}
	return sb.String()
}

// disambiguation determines if file/rank disambiguation is needed for a piece move.
func (g *Game) disambiguation(piece Piece, move Move) (needFile bool, needRank bool) {
	if piece.Type == Pawn || piece.Type == King { // King moves rarely ambiguous except castling handled earlier
		return false, false
	}
	// Find other same-type pieces that can also move to destination
	for sq := Square(0); sq < 64; sq++ {
		if sq == move.From {
			continue
		}
		p := g.board.GetPiece(sq)
		if p.IsEmpty() || p.Color != piece.Color || p.Type != piece.Type {
			continue
		}
		// Generate pseudo-legal moves for that piece
		candidates := g.generatePseudoLegalMoves(sq, p)
		for _, cand := range candidates {
			if cand.To != move.To || !g.IsLegalMove(cand) {
				continue
			}
			// Another same-type piece can also move to destination.
			// Decide minimal SAN disambiguation per FIDE rules:
			// 1. If pieces share file -> use rank.
			// 2. Else if share rank -> use file.
			// 3. Else (different file and rank) -> use file only.
			if sq.File() == move.From.File() {
				needRank = true
			} else if sq.Rank() == move.From.Rank() {
				needFile = true
			} else {
				needFile = true
			}
		}
	}
	return
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
