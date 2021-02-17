package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	uchess "github.com/tmountain/uchess/pkg"
)

func main() {
	// Game state
	var gs uchess.GameState
	// Init via flags
	gs.Config = uchess.Init()
	// Import additional themes if available
	cfgTheme, err := uchess.ImportThemes(gs.Config.ActiveTheme, gs.Config.Themes)
	// Chess board state
	gs.Game = chess.NewGame()

	// A valid theme is required. This should not happen unless someone
	// changes the config to point to something invalid
	if err != nil {
		panic(err)
	} else {
		gs.Theme = cfgTheme
	}

	// Load the FEN if applicable
	fen, err := chess.FEN(gs.Config.FEN)
	if err != nil {
		panic(err)
	}

	// This encapsulates the ongoing game state
	gs.Game = chess.NewGame(fen)
	// Connect to the UCI engines
	cfgWhite, cfgBlack, cfgHint := uchess.ImportEngines(gs.Config.UCIWhite, gs.Config.UCIBlack, gs.Config.UCIHint, gs.Config.UCIEngines)
	uciWhite, uciBlack, uciHint := uchess.InitEngines(gs.Config)
	// Store the resulting values in the game state
	// Configs are stored, to provide future opportunities for commands
	// to adjust UCI behavior on the fly
	gs.UCI.UciWhite = uciWhite
	gs.UCI.UciBlack = uciBlack
	gs.UCI.UciHint = uciHint
	gs.UCI.CfgWhite = cfgWhite
	gs.UCI.CfgBlack = cfgBlack
	gs.UCI.CfgHint = cfgHint

	// Initialize screen
	gs.S, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := gs.S.Init(); err != nil {
		panic(err)
	}
	gs.S.SetStyle(uchess.DefStyle)
	gs.S.Clear()

	// Input buffer
	gs.Input = uchess.NewInput()
	// This handles if the CPU is white (going first)
	if uchess.IsCPU(gs.Game.Position().Turn(), gs.Config) {
		// Render before the CPU goes to avoid a blank screen
		uchess.Render(&gs)
		uchess.EngMove(gs.Game, gs.UCI, gs.Config)
	}
	// Set the initial score
	gs.Score = uchess.EngScore(gs.Game, gs.UCI, gs.Config)
	uchess.Render(&gs)

	for {
		rescore := Interact(&gs)
		// If the game is still in play, update the score
		if gs.Game.Outcome() == "*" && rescore {
			gs.Score = uchess.EngScore(gs.Game, gs.UCI, gs.Config)
		}
		uchess.Render(&gs)
	}
}

// setChecks indicates if either color is in check
func setChecks(gs *uchess.GameState) {
	checkWhite, checkBlack := uchess.InCheck(gs.Game)
	gs.CheckWhite = checkWhite
	gs.CheckBlack = checkBlack
}

// Interact polls user input an dispatches appropriately
func Interact(gs *uchess.GameState) bool {
	// This is used to avoid rescoring on every key press
	rescore := true

	quit := func() {
		gs.S.Fini()
		os.Exit(0)
	}

	// False when playing cpu vs cpu
	isInteractive := uchess.IsInteractive(gs.Config)

	// Poll event
	ev := gs.S.PollEvent()

	// Message
	msg := ""

	// Process event
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		// Quit the app
		case tcell.KeyEscape, tcell.KeyCtrlC:
			quit()
		// Redraw
		case tcell.KeyCtrlL:
			gs.S.Sync()
		case tcell.KeyEnter:
			if isInteractive {
				// Reset hints when new commands come through
				gs.Hint = nil
				// Attempt to process the command
				msg, gs.Game = uchess.ProcessCmd(gs.Input.Current(), gs)
				// Set the check state in the event that a check happened
				setChecks(gs)
				uchess.DrawMsgLabel(gs.S, msg, gs.Theme)
				gs.Input.Clear()
				// Render between moves to show the square that was chosen
				uchess.Render(gs)
			}

			// If the game is still in play, CPU takes move (if applicable)
			if gs.Game.Outcome() == "*" && uchess.IsCPU(gs.Game.Position().Turn(), gs.Config) {
				uchess.DrawMsgLabel(gs.S, "Thinking...", gs.Theme)
				// Render between moves to show the square that was chosen
				uchess.Render(gs)
				msg = uchess.EngMove(gs.Game, gs.UCI, gs.Config)
				// Set the check state in the event that a check happened
				setChecks(gs)
				uchess.DrawMsgLabel(gs.S, msg, gs.Theme)
			}
		// Backspace
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if isInteractive {
				rescore = false
				gs.Input.Backspace()
			}
		// Append input
		default:
			if isInteractive {
				rescore = false
				gs.Input.Append(ev.Rune())
			}
		}

	case *tcell.EventResize:
		gs.S.Sync()
	}
	return rescore
}
