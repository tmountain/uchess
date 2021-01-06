package uchess

import (
	"fmt"
	"os"
	"strings"

	"github.com/freeeve/uci"
	"github.com/notnil/chess"
	"github.com/notnil/chess/image"
)

// selectEngine returns the UCI engine and its corresponding config based upon the current turn
func selectEngine(game *chess.Game, us UCIState) (*uci.Engine, *UCIEngine) {
	var eng *uci.Engine
	var cfg *UCIEngine

	if game.Position().Turn() == chess.White {
		eng = us.UciWhite
		cfg = us.CfgWhite
	} else {
		eng = us.UciBlack
		cfg = us.CfgBlack
	}
	return eng, cfg
}

// EngScore provides the current board score in centipawns for whomever the current
// game position identifies as active
func EngScore(game *chess.Game, us UCIState, config Config) int {
	eng, _ := selectEngine(game, us)
	// Set some result filter options
	resultOpts := uci.HighestDepthOnly | uci.IncludeUpperbounds | uci.IncludeLowerbounds
	// Update the engine on the game position
	eng.SetFEN(game.Position().String())
	// Do a quick analysis of the current board
	results, _ := eng.GoDepth(10, resultOpts)
	resultLen := len(results.Results)
	// Some super minimal UCI engines don't report the score
	// This prevents a crash
	if resultLen > 0 {
		return results.Results[resultLen-1].Score
	}
	// Default to reporting 0 for lack of a better option
	return 0
}

// EngMove executes a move using the UCI engine
func EngMove(game *chess.Game, us UCIState, config Config) string {
	eng, engCfg := selectEngine(game, us)

	// Update the engine on the game position
	eng.SetFEN(game.Position().String())
	// Set some result filter options
	resultOpts := uci.HighestDepthOnly | uci.IncludeUpperbounds | uci.IncludeLowerbounds
	results, _ := eng.Go(engCfg.Depth, engCfg.SearchMoves, engCfg.MoveTime, resultOpts)
	// Take the best move
	bestMove := results.BestMove
	// Convert the game position to long algebraic notation
	move, err := chess.LongAlgebraicNotation{}.Decode(game.Position(), bestMove)
	// This should never happen unless the UCI returns invalid notation
	if err != nil {
		return "\u26A0 Error. Engine notation."
	}

	// This should never happen unless the UCI returns an invalid move
	if err = game.Move(move); err != nil {
		return "\u26A0 Error. Engine move."
	}

	// Clear the label
	return strings.Repeat(" ", 32)
}

func undoMove(game *chess.Game) *chess.Game {
	newGame := chess.NewGame()
	moves := game.Moves()
	for i := 0; i < len(moves)-2; i++ {
		move := moves[i]
		newGame.Move(move)
	}
	return newGame
}

func resetGame(game *chess.Game) *chess.Game {
	newGame := chess.NewGame()
	return newGame
}

func saveGame(game *chess.Game) string {
	ts := Timestamp()
	file := fmt.Sprintf("uchess_%v.txt", ts)
	f, err := os.Create(file)
	defer f.Close()

	if err != nil {
		return err.Error()
	}
	f.WriteString(game.String())
	return fmt.Sprintf("Saved %v", file)
}

func saveImage(game *chess.Game) string {
	ts := Timestamp()
	file := fmt.Sprintf("uchess_%v.svg", ts)
	f, err := os.Create(file)
	defer f.Close()

	if err != nil {
		return err.Error()
	}

	board := game.Position().Board()
	if err := image.SVG(f, board); err != nil {
		return err.Error()
	}
	return fmt.Sprintf("Saved %v", file)
}

func resign(game *chess.Game) *chess.Game {
	game.Resign(game.Position().Turn())
	return game
}

func hint(gs *GameState) string {
	DrawMsgLabel(gs.S, "Thinking...", gs.Theme)
	Render(gs)
	eng := gs.UCI.UciHint
	engCfg := gs.UCI.CfgHint
	// Update the engine on the game position
	eng.SetFEN(gs.Game.Position().String())
	// Set some result filter options
	resultOpts := uci.HighestDepthOnly | uci.IncludeUpperbounds | uci.IncludeLowerbounds
	results, _ := eng.Go(engCfg.Depth, engCfg.SearchMoves, engCfg.MoveTime, resultOpts)
	// Take the best move
	bestMove := results.BestMove
	// Convert the game position to long algebraic notation
	move, err := chess.LongAlgebraicNotation{}.Decode(gs.Game.Position(), bestMove)
	// This should never happen unless the UCI returns invalid notation
	if err != nil {
		return "\u26A0 Error. Engine notation."
	}

	// Success, set the move in the game state
	gs.Hint = move
	return strings.Repeat(" ", 80)
}

// ProcessCmd processes a move request or command
func ProcessCmd(cmd string, gs *GameState) (string, *chess.Game) {
	cmd = strings.TrimSpace(cmd)

	switch cmd {
	// Back one turn
	case "back":
		return strings.Repeat(" ", 80), undoMove(gs.Game)
		// Save the PGN string
	case "save":
		return saveGame(gs.Game), gs.Game
		// SVG snapshot of the current board
	case "image":
		return saveImage(gs.Game), gs.Game
		// Output the FEN string
	case "fen":
		return gs.Game.Position().String(), gs.Game
		// Reset the game
	case "reset":
		return strings.Repeat(" ", 80), resetGame(gs.Game)
		// Current player resigns
	case "resign":
		return strings.Repeat(" ", 80), resign(gs.Game)
		// Process a move string
	case "hint":
		return hint(gs), gs.Game
	default:
		if err := gs.Game.MoveStr(cmd); err != nil {
			return "\u26A0 Illegal. Try again.", gs.Game
		}
	}

	// Clear the label
	return strings.Repeat(" ", 80), gs.Game
}
