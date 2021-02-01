package uchess

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/markbates/pkger"
)

// Config defines the configurable parameters for UCI
type Config struct {
	UCIWhite    string      `json:"uciWhite"`
	UCIBlack    string      `json:"uciBlack"`
	UCIHint     string      `json:"uciHint"`
	UCIEngines  []UCIEngine `json:"uciEngines"`
	FEN         string      `json:"fen"`
	ActiveTheme string      `json:"activeTheme"`
	Themes      []ThemeHex  `json:"theme"`
	WhitePiece  string      `json:"whitePiece"`
	BlackPiece  string      `json:"blackPiece"`
	WhiteName   string      `json:"whiteName"`
	BlackName   string      `json:"blackName"`
}

// HasTheme returns a bool indicating whether the config
// contains a theme of the name provided
func HasTheme(name string, themes []ThemeHex) bool {
	for _, theme := range themes {
		if theme.Name == name {
			return true
		}
	}
	return false
}

// DefaultOptions defines any default UCI options
var DefaultOptions = []Option{
	{"skill level", "20"}, // stockfish skill level
}

var uciEngines = []UCIEngine{
	{
		"stockfish",            // Name
		FindOrFetchStockfish(), // Path
		128,                    // Hash
		false,                  // Ponder
		false,                  // OwnBook
		4,                      // MultiPV
		0,                      // Depth
		"",                     // SearchMoves
		3000,                   // MoveTime
		DefaultOptions,
	},
}

// defaultFEN is the default board position
// The format is Forsyth-Edwards notation
// https://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
const defaultFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// ReadThemes reads packaged theme data into a ThemeHex data slice
func ReadThemes() []ThemeHex {
	var themes []ThemeHex
	pkger.Walk("/themes", func(path string, info os.FileInfo, err error) error {
		var theme ThemeHex
		f, err := pkger.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if !info.IsDir() {
			msgBytes, _ := ioutil.ReadAll(f)
			json.Unmarshal(msgBytes, &theme)
			themes = append(themes, theme)
		}
		return nil
	})
	return themes
}

// DefaultConfig defines the default configuration
var DefaultConfig = Config{
	"stockfish",  // UCIWhite
	"stockfish",  // UCIBlack
	"stockfish",  // UCIHint
	uciEngines,   // UCIEngine
	defaultFEN,   // FEN
	"basic",      // ActiveTheme
	[]ThemeHex{}, // Themes
	"human",      // WhitePiece
	"cpu",        // BlackPiece
	"",           // WhiteName
	"",           // BlackName
}

// ConfigJSON returns the JSON encoded representation of the config
func ConfigJSON() string {
	config := DefaultConfig
	// Just include the basic theme in the default config as a reference
	config.Themes = []ThemeHex{ThemeBasic.Hex()}

	c, err := json.MarshalIndent(&DefaultConfig, "", "    ")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v", string(c))
}

// ReadConfig reads the specified JSON file into a config struct
func ReadConfig(file string) Config {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	allThemes := make([]ThemeHex, 0)
	// Load builtin themes while allowing
	// user defined themes to override builtin
	for _, theme := range ReadThemes() {
		if !HasTheme(theme.Name, config.Themes) {
			allThemes = append(allThemes, theme)
		}
	}

	// Load user defind themes from the config
	for _, theme := range config.Themes {
		allThemes = append(allThemes, theme)
	}
	config.Themes = allThemes
	return config
}
