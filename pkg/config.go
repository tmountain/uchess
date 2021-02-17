package uchess

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
)

//go:embed themes/*
var content embed.FS

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
	{"skill level", "3"}, // stockfish skill level
}

// DefaultEngine is a placeholder for the default UCI engine
var defaultEngine = UCIEngine{
	"stockfish", // Name
	"",          // Path to UCI (filled in dynamically)
	128,         // Hash
	false,       // Ponder
	false,       // OwnBook
	1,           // MultiPV
	1,           // Depth
	"",          // SearchMoves
	100,         // MoveTime
	DefaultOptions,
}

// defaultFEN is the default board position
// The format is Forsyth-Edwards notation
// https://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
const defaultFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// ReadThemes reads packaged theme data into a ThemeHex data slice
func ReadThemes() []ThemeHex {
	var themes []ThemeHex
	var theme ThemeHex
	files, err := content.ReadDir("themes")

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		bytes, err := content.ReadFile(path.Join("themes", file.Name()))
		if err != nil {
			panic(err)
		}
		json.Unmarshal(bytes, &theme)
		themes = append(themes, theme)
	}

	return themes
}

// DefaultConfig defines the default configuration
var defaultConfig = Config{
	"stockfish",   // UCIWhite
	"stockfish",   // UCIBlack
	"stockfish",   // UCIHint
	[]UCIEngine{}, // UCIEngine
	defaultFEN,    // FEN
	"basic",       // ActiveTheme
	[]ThemeHex{},  // Themes
	"human",       // WhitePiece
	"cpu",         // BlackPiece
	"",            // WhiteName
	"",            // BlackName
}

// MakeDefault creates the default config
func MakeDefault() Config {
	config := defaultConfig
	engine := defaultEngine
	// Try to find stockfish in the path
	stockfish := FindStockfish()

	// If stockfish cannot be found, set the config to the AppDir
	if stockfish == "" {
		stockfish = filepath.Join(AppDir(), StockfishFilename())
	}

	engine.Path = stockfish
	config.UCIEngines = []UCIEngine{engine}
	return config
}

// ConfigJSON returns the JSON encoded representation of the config
func ConfigJSON(config Config) string {
	// Just include the basic theme in the default config as a reference
	config.Themes = []ThemeHex{ThemeBasic.Hex()}

	c, err := json.MarshalIndent(&config, "", "    ")
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

	// Load user defined themes from the config
	for _, theme := range config.Themes {
		allThemes = append(allThemes, theme)
	}
	config.Themes = allThemes
	return config
}
