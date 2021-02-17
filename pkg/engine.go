package uchess

import (
	"fmt"
	"time"

	"github.com/notnil/chess/uci"
)

// Option allows arbitrary UCI options to be sent
type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// UCIEngine defines a UCIEngine configuration
type UCIEngine struct {
	Name        string        `json:"name"`
	Path        string        `json:"engine"`
	Hash        int           `json:"hash"`
	Ponder      bool          `json:"ponder"`
	OwnBook     bool          `json:"ownBook"`
	MultiPV     int           `json:"multiPV"`
	Depth       int           `json:"depth"`
	SearchMoves string        `json:"searchMoves"`
	MoveTime    time.Duration `json:"moveTime"`
	Options     []Option      `json:"options"`
}

// UCIState holds the UCI engine state
type UCIState struct {
	UciWhite *uci.Engine // UCI Engine
	UciBlack *uci.Engine // UCI Engine
	UciHint  *uci.Engine // UCI Engine
	CfgWhite *UCIEngine  // UCI Engine Config
	CfgBlack *UCIEngine  // UCI Engine Config
	CfgHint  *UCIEngine  // UCI Engine Config
}

// cfgEngines configures the UCI engine
func cfgEngines(eng *uci.Engine, cfg *UCIEngine) {
	optHash := uci.CmdSetOption{Name: "hash", Value: fmt.Sprintf("%v", cfg.Hash)}
	optPonder := uci.CmdSetOption{Name: "ponder", Value: fmt.Sprintf("%v", cfg.Ponder)}
	optOwnBook := uci.CmdSetOption{Name: "ownbook", Value: fmt.Sprintf("%v", cfg.OwnBook)}
	optMultiPV := uci.CmdSetOption{Name: "multipv", Value: fmt.Sprintf("%v", cfg.MultiPV)}

	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, optHash, optPonder, optOwnBook, optMultiPV, uci.CmdUCINewGame); err != nil {
		panic(err)
	}

	// Send any custom options that were specified
	for _, option := range cfg.Options {
		uci.CmdSetOption{Name: option.Name, Value: option.Value}.ProcessResponse(eng)
	}
}

// InitEngines configures UCI engines if the config dictates that
// they are required. The white and black engines are returned
// respectively
func InitEngines(config Config) (*uci.Engine, *uci.Engine, *uci.Engine) {
	cfgWhite, cfgBlack, cfgHint := ImportEngines(config.UCIWhite, config.UCIBlack, config.UCIHint, config.UCIEngines)
	var engWhite, engBlack, engHint *uci.Engine

	engWhite, err := uci.New(cfgWhite.Path)
	if err != nil {
		panic(err)
	}
	cfgEngines(engWhite, cfgWhite)

	engBlack, err = uci.New(cfgBlack.Path)
	if err != nil {
		panic(err)
	}
	cfgEngines(engBlack, cfgBlack)

	engHint, err = uci.New(cfgHint.Path)
	if err != nil {
		panic(err)
	}
	cfgEngines(engHint, cfgHint)

	return engWhite, engBlack, engHint
}

// ImportEngines returns a UCIEngine config for white and black
func ImportEngines(uciWhite string, uciBlack string, uciHint string, engines []UCIEngine) (*UCIEngine, *UCIEngine, *UCIEngine) {
	var cfgWhite, cfgBlack, cfgHint UCIEngine
	var whiteFound, blackFound, hintFound bool

	for _, e := range engines {
		if uciWhite == e.Name {
			cfgWhite = e
			whiteFound = true
		}

		if uciBlack == e.Name {
			cfgBlack = e
			blackFound = true
		}

		if uciHint == e.Name {
			cfgHint = e
			hintFound = true
		}
	}

	if !whiteFound {
		panic("failed to import white engine config")
	}

	if !blackFound {
		panic("failed to import black engine config")
	}

	if !hintFound {
		panic("failed to import hint engine config")
	}

	return &cfgWhite, &cfgBlack, &cfgHint
}
