package main

import (
	"fmt"

	"github.com/rumendamyanov/go-chess/engine"
)

func main() {
	// Create a game and make a move
	game := engine.NewGame()

	// Make e2e4
	move, err := game.ParseMove("e2e4")
	if err != nil {
		panic(err)
	}

	err = game.MakeMove(move)
	if err != nil {
		panic(err)
	}

	// Print the FEN
	fmt.Printf("FEN after e2e4: %s\n", game.ToFEN())

	// Get legal moves
	legalMoves := game.GetAllLegalMoves()
	fmt.Printf("Number of legal moves: %d\n", len(legalMoves))
	fmt.Printf("First few legal moves: ")
	for i, move := range legalMoves[:5] {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(move.String())
	}
	fmt.Println()

	// Check game status
	fmt.Printf("Game status: %s\n", game.Status().String())
	fmt.Printf("Active color: %s\n", game.ActiveColor().String())
	fmt.Printf("In check: %v\n", game.Status() == engine.Check)
}
