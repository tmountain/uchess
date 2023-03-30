package uchess

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/image"
	"github.com/notnil/chess/uci"
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
	// Update the engine on the game position
	cmdPos := uci.CmdPosition{Position: game.Position()}
	// Do a quick analysis of the current board
	cmdGo := uci.CmdGo{Depth: 10}
	// If the command fails for some reason, exit
	if err := eng.Run(cmdPos, cmdGo); err != nil {
		panic(err)
	}
	info := eng.SearchResults().Info
	// Return the score in centipawns
	return info.Score.CP
}

// EngMove executes a move using the UCI engine
func EngMove(game *chess.Game, us UCIState, config Config) string {
	eng, engCfg := selectEngine(game, us)

	// Update the engine on the game position
	cmdPos := uci.CmdPosition{Position: game.Position()}
	// Params to send to the engine
	cmdGo := &uci.CmdGo{Depth: engCfg.Depth}
	cmdGo.MoveTime = engCfg.MoveTime * time.Millisecond
	// If SearchMoves is specified, include it
	if engCfg.SearchMoves != "" {
		cmdGo.SearchMoves = searchMoves(engCfg.SearchMoves)
	}
	// Run the actual commands
	if err := eng.Run(cmdPos, *cmdGo); err != nil {
		return "\u26A0 Error. Engine command."
	}
	// Fetch the results
	move := eng.SearchResults().BestMove
	// Validate the move
	if err := game.Move(move); err != nil {
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
	cmdPos := uci.CmdPosition{Position: gs.Game.Position()}
	// Params to send to the engine
	cmdGo := &uci.CmdGo{Depth: engCfg.Depth}
	cmdGo.MoveTime = engCfg.MoveTime * time.Millisecond
	// If SearchMoves is specified, include it
	if engCfg.SearchMoves != "" {
		cmdGo.SearchMoves = searchMoves(engCfg.SearchMoves)
	}
	// Run the actual commands
	if err := eng.Run(cmdPos, *cmdGo); err != nil {
		return "\u26A0 Error. Engine command."
	}
	// Take the best move
	bestMove := eng.SearchResults().BestMove
	// Success, set the move in the game state
	gs.Hint = bestMove
	return strings.Repeat(" ", 80)
}

func quit(gs *GameState) string {
	return "quit"
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
	case "quit":
		return quit(gs), gs.Game
	default:
		if err := gs.Game.MoveStr(cmd); err != nil {
			return "\u26A0 Illegal. Try again.", gs.Game
		}
	}

	// Clear the label
	return strings.Repeat(" ", 80), gs.Game
}
