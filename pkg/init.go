package uchess

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// uciCmd returns the command name from the UCI command path
func uciCmd(p string) string {
	_, file := filepath.Split(p)
	return fmt.Sprintf("%v", file)
}

// osUser returns the username for the currently logged in user
func osUser() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.Username
}

// setBlackPieceName sets the player name for the black pieces
func setBlackPieceName(c *Config) {
	if c.BlackPiece == "cpu" {
		c.BlackName = uciCmd(c.UCIBlack)
	} else {
		c.BlackName = osUser()
	}
}

// setWhitePieceName sets the player name for the white pieces
func setWhitePieceName(c *Config) {
	if c.WhitePiece == "cpu" {
		c.WhiteName = uciCmd(c.UCIWhite)
	} else {
		c.WhiteName = osUser()
	}
}

// Init sets up the app config
func Init() Config {
	var config Config
	tmpl := flag.Bool("tmpl", false, "generate config template with defaults")
	cfg := flag.String("cfg", "", "config file")
	white := flag.String("white", "human", "white piece input")
	black := flag.String("black", "cpu", "black piece input")
	themes := flag.Bool("themes", false, "list theme names and exit")

	flag.Parse()

	if *tmpl {
		defaultCfg := MakeDefault()
		fmt.Println(ConfigJSON(defaultCfg))
		os.Exit(0)
	}

	if *themes {
		themes := ReadThemes()
		for _, theme := range themes {
			fmt.Println(theme.Name)
		}
		os.Exit(0)
	}

	if (*white != "human" && *white != "cpu") ||
		(*black != "human" && *black != "cpu") {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Load config file if provided
	// Override command line flags for black/white
	// If UCI engine is specified in config, it is taken at face value
	if *cfg != "" {
		config = ReadConfig(*cfg)
		*white = config.WhitePiece
		*black = config.BlackPiece
	} else {
		// Zero configuration config (hopefully)
		config = MakeDefault()
		config.Themes = ReadThemes()
		uciPath := config.UCIEngines[0].Path

		// If the UCI engine cannot be found, prompt to install if applicable
		if !IsFile(uciPath) {
			FetchStockfish()
		}
	}

	config.WhitePiece = *white
	config.BlackPiece = *black

	if config.WhiteName == "" {
		setWhitePieceName(&config)
	}

	if config.BlackName == "" {
		setBlackPieceName(&config)
	}

	return config
}
