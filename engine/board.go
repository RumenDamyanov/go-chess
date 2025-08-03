// Package engine provides the core chess engine implementation.
// It includes board representation, move generation, game state management,
// and rule validation according to standard chess rules.
package engine

import (
	"fmt"
	"strings"
)

// Color represents the color of a chess piece or player.
type Color int

const (
	// None represents no color (empty squares).
	None Color = iota
	// White represents the white player.
	White
	// Black represents the black player.
	Black
)

// String returns the string representation of a color.
func (c Color) String() string {
	switch c {
	case None:
		return "none"
	case White:
		return "white"
	case Black:
		return "black"
	default:
		return "unknown"
	}
}

// PieceType represents the type of a chess piece.
type PieceType int

const (
	// Empty represents an empty square.
	Empty PieceType = iota
	// Pawn represents a pawn piece.
	Pawn
	// Rook represents a rook piece.
	Rook
	// Knight represents a knight piece.
	Knight
	// Bishop represents a bishop piece.
	Bishop
	// Queen represents a queen piece.
	Queen
	// King represents a king piece.
	King
)

// String returns the string representation of a piece type.
func (pt PieceType) String() string {
	switch pt {
	case Empty:
		return "empty"
	case Pawn:
		return "pawn"
	case Rook:
		return "rook"
	case Knight:
		return "knight"
	case Bishop:
		return "bishop"
	case Queen:
		return "queen"
	case King:
		return "king"
	default:
		return "unknown"
	}
}

// Piece represents a chess piece with its type and color.
type Piece struct {
	Type  PieceType
	Color Color
}

// IsEmpty returns true if the piece represents an empty square.
func (p Piece) IsEmpty() bool {
	return p.Type == Empty
}

// String returns the string representation of a piece.
func (p Piece) String() string {
	if p.IsEmpty() {
		return "."
	}

	symbol := ""
	switch p.Type {
	case Pawn:
		symbol = "P"
	case Rook:
		symbol = "R"
	case Knight:
		symbol = "N"
	case Bishop:
		symbol = "B"
	case Queen:
		symbol = "Q"
	case King:
		symbol = "K"
	}

	if p.Color == Black {
		symbol = strings.ToLower(symbol)
	}

	return symbol
}

// Square represents a position on the chess board.
type Square int

const (
	// A1 through H8 represent the 64 squares of a chess board.
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
)

// SquareFromString parses a square from algebraic notation (e.g., "e4").
func SquareFromString(s string) (Square, error) {
	if len(s) != 2 {
		return 0, fmt.Errorf("invalid square notation: %s", s)
	}

	file := s[0] - 'a'
	rank := s[1] - '1'

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return 0, fmt.Errorf("invalid square notation: %s", s)
	}

	return Square(int(rank)*8 + int(file)), nil
}

// String returns the algebraic notation of the square (e.g., "e4").
func (s Square) String() string {
	if s < 0 || s > 63 {
		return "invalid"
	}

	file := s % 8
	rank := s / 8

	return fmt.Sprintf("%c%c", 'a'+file, '1'+rank)
}

// File returns the file (column) of the square (0-7).
func (s Square) File() int {
	return int(s % 8)
}

// Rank returns the rank (row) of the square (0-7).
func (s Square) Rank() int {
	return int(s / 8)
}

// Board represents a chess board with piece positions.
type Board struct {
	squares [64]Piece
}

// NewBoard creates a new board with the standard starting position.
func NewBoard() *Board {
	b := &Board{}
	b.SetupStartingPosition()
	return b
}

// SetupStartingPosition sets up the board with the standard chess starting position.
func (b *Board) SetupStartingPosition() {
	// Clear the board
	for i := range b.squares {
		b.squares[i] = Piece{Type: Empty}
	}

	// Place white pieces
	b.squares[A1] = Piece{Type: Rook, Color: White}
	b.squares[B1] = Piece{Type: Knight, Color: White}
	b.squares[C1] = Piece{Type: Bishop, Color: White}
	b.squares[D1] = Piece{Type: Queen, Color: White}
	b.squares[E1] = Piece{Type: King, Color: White}
	b.squares[F1] = Piece{Type: Bishop, Color: White}
	b.squares[G1] = Piece{Type: Knight, Color: White}
	b.squares[H1] = Piece{Type: Rook, Color: White}

	for i := A2; i <= H2; i++ {
		b.squares[i] = Piece{Type: Pawn, Color: White}
	}

	// Place black pieces
	b.squares[A8] = Piece{Type: Rook, Color: Black}
	b.squares[B8] = Piece{Type: Knight, Color: Black}
	b.squares[C8] = Piece{Type: Bishop, Color: Black}
	b.squares[D8] = Piece{Type: Queen, Color: Black}
	b.squares[E8] = Piece{Type: King, Color: Black}
	b.squares[F8] = Piece{Type: Bishop, Color: Black}
	b.squares[G8] = Piece{Type: Knight, Color: Black}
	b.squares[H8] = Piece{Type: Rook, Color: Black}

	for i := A7; i <= H7; i++ {
		b.squares[i] = Piece{Type: Pawn, Color: Black}
	}
}

// GetPiece returns the piece at the given square.
func (b *Board) GetPiece(sq Square) Piece {
	if sq < 0 || sq > 63 {
		return Piece{Type: Empty}
	}
	return b.squares[sq]
}

// SetPiece sets the piece at the given square.
func (b *Board) SetPiece(sq Square, piece Piece) {
	if sq >= 0 && sq <= 63 {
		b.squares[sq] = piece
	}
}

// String returns a string representation of the board.
func (b *Board) String() string {
	var sb strings.Builder

	sb.WriteString("  a b c d e f g h\n")

	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(fmt.Sprintf("%d ", rank+1))
		for file := 0; file < 8; file++ {
			square := Square(rank*8 + file)
			piece := b.GetPiece(square)
			sb.WriteString(piece.String())
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%d\n", rank+1))
	}

	sb.WriteString("  a b c d e f g h\n")

	return sb.String()
}

// Copy returns a deep copy of the board.
func (b *Board) Copy() *Board {
	newBoard := &Board{}
	copy(newBoard.squares[:], b.squares[:])
	return newBoard
}
