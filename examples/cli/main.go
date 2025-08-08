package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"go.rumenx.com/chess/ai"
	"go.rumenx.com/chess/engine"
)

func main() {
	fmt.Println("Welcome to go-chess CLI!")
	fmt.Println("Type 'help' for commands, 'quit' to exit")
	fmt.Println()

	// Create a new game
	game := engine.NewGame()

	// Create AI opponent
	aiPlayer := ai.NewMinimaxAI(ai.DifficultyMedium)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Starting position:")
	fmt.Println(game.Board().String())
	fmt.Printf("%s to move. Enter your move (e.g., 'e2e4'): ", game.ActiveColor().String())

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			fmt.Println("Thanks for playing!")
			return
		case "help", "h":
			printHelp()
		case "board", "b":
			fmt.Println(game.Board().String())
		case "status", "s":
			printGameStatus(game)
		case "history":
			printMoveHistory(game)
		case "new":
			game = engine.NewGame()
			fmt.Println("New game started!")
			fmt.Println(game.Board().String())
		default:
			if input == "" {
				// Empty input, just continue
			} else {
				// Try to parse as a move
				if err := handleMove(game, input); err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					// Check game status
					if game.Status() != engine.InProgress {
						fmt.Printf("Game over! Status: %s\n", game.Status().String())
						fmt.Print("Type 'new' to start a new game or 'quit' to exit: ")
						continue
					}

					// AI move
					if game.ActiveColor() == engine.Black {
						fmt.Println("AI is thinking...")
						ctx := context.Background()
						aiMove, err := aiPlayer.GetBestMove(ctx, game)
						if err != nil {
							fmt.Printf("AI error: %v\n", err)
						} else {
							if err := game.MakeMove(aiMove); err != nil {
								fmt.Printf("AI move error: %v\n", err)
							} else {
								fmt.Printf("AI plays: %s\n", aiMove.String())
								fmt.Println(game.Board().String())

								// Check game status again
								if game.Status() != engine.InProgress {
									fmt.Printf("Game over! Status: %s\n", game.Status().String())
									fmt.Print("Type 'new' to start a new game or 'quit' to exit: ")
									continue
								}
							}
						}
					}

					fmt.Println(game.Board().String())
				}
			}
		}

		if game.Status() == engine.InProgress {
			fmt.Printf("%s to move. Enter your move: ", game.ActiveColor().String())
		} else {
			fmt.Print("Enter command: ")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading input:", err)
	}
}

func handleMove(game *engine.Game, input string) error {
	move, err := game.ParseMove(input)
	if err != nil {
		return fmt.Errorf("invalid move notation: %v", err)
	}

	if err := game.MakeMove(move); err != nil {
		return fmt.Errorf("illegal move: %v", err)
	}

	fmt.Printf("Move played: %s\n", move.String())
	return nil
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  help, h      - Show this help")
	fmt.Println("  board, b     - Show current board")
	fmt.Println("  status, s    - Show game status")
	fmt.Println("  history      - Show move history")
	fmt.Println("  new          - Start new game")
	fmt.Println("  quit, exit, q - Quit the game")
	fmt.Println()
	fmt.Println("Move notation:")
	fmt.Println("  e2e4         - Move from e2 to e4")
	fmt.Println("  e7e8Q        - Pawn promotion to Queen")
	fmt.Println("  O-O          - Kingside castling")
	fmt.Println("  O-O-O        - Queenside castling")
	fmt.Println()
}

func printGameStatus(game *engine.Game) {
	fmt.Printf("Status: %s\n", game.Status().String())
	fmt.Printf("Active color: %s\n", game.ActiveColor().String())
	fmt.Printf("Move count: %d\n", game.MoveCount())
	fmt.Printf("Moves played: %d\n", len(game.MoveHistory()))
}

func printMoveHistory(game *engine.Game) {
	history := game.MoveHistory()
	if len(history) == 0 {
		fmt.Println("No moves played yet.")
		return
	}

	fmt.Println("Move history:")
	for i, move := range history {
		color := "White"
		if i%2 == 1 {
			color = "Black"
		}
		fmt.Printf("  %d. %s: %s\n", i/2+1, color, move.String())
	}
}
