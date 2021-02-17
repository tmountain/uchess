package uchess

import (
	"errors"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Terminal safe color palette is available here
// Themes should be limited to the colors defined in this reference
// https://upload.wikimedia.org/wikipedia/commons/1/15/Xterm_256color_chart.svg

// Theme is used for dynamically coloring the UI
type Theme struct {
	Name         string      `json:"name"`
	MoveLabelBg  tcell.Color `json:"moveLabelBg"`
	MoveLabelFg  tcell.Color `json:"moveLabelFg"`
	SquareDark   tcell.Color `json:"squareDark"`
	SquareLight  tcell.Color `json:"squareLight"`
	SquareHigh   tcell.Color `json:"squareHigh"`
	SquareHint   tcell.Color `json:"squareHint"`
	SquareCheck  tcell.Color `json:"squareCheck"`
	White        tcell.Color `json:"white"`
	Black        tcell.Color `json:"black"`
	Msg          tcell.Color `json:"msg"`
	Rank         tcell.Color `json:"rank"`
	File         tcell.Color `json:"file"`
	Prompt       tcell.Color `json:"prompt"`
	MeterBase    tcell.Color `json:"meterBase"`
	MeterMid     tcell.Color `json:"meterMid"`
	MeterNeutral tcell.Color `json:"meterNeutral"`
	MeterWin     tcell.Color `json:"meterWin"`
	MeterLose    tcell.Color `json:"meterLose"`
	PlayerNames  tcell.Color `json:"playerNames"`
	Score        tcell.Color `json:"score"`
	MoveBox      tcell.Color `json:"moveBox"`
	Emoji        tcell.Color `json:"emoji"`
	Input        tcell.Color `json:"input"`
	Advantage    tcell.Color `json:"advantage"`
}

// ThemeHex is used for dynamically coloring the UI
type ThemeHex struct {
	Name         string `json:"name"`
	MoveLabelBg  string `json:"moveLabelBg"`
	MoveLabelFg  string `json:"moveLabelFg"`
	SquareDark   string `json:"squareDark"`
	SquareLight  string `json:"squareLight"`
	SquareHigh   string `json:"squareHigh"`
	SquareHint   string `json:"squareHint"`
	SquareCheck  string `json:"squareCheck"`
	White        string `json:"white"`
	Black        string `json:"black"`
	Msg          string `json:"msg"`
	Rank         string `json:"rank"`
	File         string `json:"file"`
	Prompt       string `json:"prompt"`
	MeterBase    string `json:"meterBase"`
	MeterMid     string `json:"meterMid"`
	MeterNeutral string `json:"meterNeutral"`
	MeterWin     string `json:"meterWin"`
	MeterLose    string `json:"meterLose"`
	PlayerNames  string `json:"playerNames"`
	Score        string `json:"score"`
	MoveBox      string `json:"moveBox"`
	Emoji        string `json:"emoji"`
	Input        string `json:"input"`
	Advantage    string `json:"advantage"`
}

// fmtHex returns a one character hex for the ColorDefault
// and otherwise it returns a standard hex. This is useful
// because it allows ColorDefault to be imported from the config
// and parsed properly rather than being interpreted as black
func fmtHex(v int32) string {
	if v == -1 {
		return "#0"
	}
	return fmt.Sprintf("#%06x", v)
}

// Hex converts a Theme to a ThemeHex
func (t Theme) Hex() ThemeHex {
	return ThemeHex{
		t.Name,
		fmtHex(t.MoveLabelBg.Hex()),
		fmtHex(t.MoveLabelFg.Hex()),
		fmtHex(t.SquareDark.Hex()),
		fmtHex(t.SquareLight.Hex()),
		fmtHex(t.SquareHigh.Hex()),
		fmtHex(t.SquareHint.Hex()),
		fmtHex(t.SquareCheck.Hex()),
		fmtHex(t.White.Hex()),
		fmtHex(t.Black.Hex()),
		fmtHex(t.Msg.Hex()),
		fmtHex(t.Rank.Hex()),
		fmtHex(t.File.Hex()),
		fmtHex(t.Prompt.Hex()),
		fmtHex(t.MeterBase.Hex()),
		fmtHex(t.MeterMid.Hex()),
		fmtHex(t.MeterNeutral.Hex()),
		fmtHex(t.MeterWin.Hex()),
		fmtHex(t.MeterLose.Hex()),
		fmtHex(t.PlayerNames.Hex()),
		fmtHex(t.Score.Hex()),
		fmtHex(t.MoveBox.Hex()),
		fmtHex(t.Emoji.Hex()),
		fmtHex(t.Input.Hex()),
		fmtHex(t.Advantage.Hex()),
	}
}

// Theme converts a ThemeHex to a Theme
func (t ThemeHex) Theme() Theme {
	return Theme{
		t.Name,
		tcell.GetColor(t.MoveLabelBg),
		tcell.GetColor(t.MoveLabelFg),
		tcell.GetColor(t.SquareDark),
		tcell.GetColor(t.SquareLight),
		tcell.GetColor(t.SquareHigh),
		tcell.GetColor(t.SquareHint),
		tcell.GetColor(t.SquareCheck),
		tcell.GetColor(t.White),
		tcell.GetColor(t.Black),
		tcell.GetColor(t.Msg),
		tcell.GetColor(t.Rank),
		tcell.GetColor(t.File),
		tcell.GetColor(t.Prompt),
		tcell.GetColor(t.MeterBase),
		tcell.GetColor(t.MeterMid),
		tcell.GetColor(t.MeterNeutral),
		tcell.GetColor(t.MeterWin),
		tcell.GetColor(t.MeterLose),
		tcell.GetColor(t.PlayerNames),
		tcell.GetColor(t.Score),
		tcell.GetColor(t.MoveBox),
		tcell.GetColor(t.Emoji),
		tcell.GetColor(t.Input),
		tcell.GetColor(t.Advantage),
	}
}

// ImportThemes returns a converted Theme from a slice of ThemeHex
// entities if its name matches the want argument
func ImportThemes(want string, themes []ThemeHex) (Theme, error) {
	// First check if want is in the provided config (override)
	for _, t := range themes {
		if t.Name == want {
			return t.Theme(), nil
		}
	}

	return Theme{}, errors.New("theme: no theme found")
}

// ThemeBasic is the default theme
var ThemeBasic = Theme{
	"basic",            // Name
	tcell.Color252,     // MoveLabelBg
	tcell.ColorBlack,   // MoveLabelFg
	tcell.Color188,     // SquareDark
	tcell.Color230,     // SquareLight
	tcell.Color226,     // SquareHigh
	tcell.Color223,     // SquareHint
	tcell.Color218,     // SquareCheck
	tcell.Color232,     // White
	tcell.Color232,     // Black
	tcell.Color160,     // Msg
	tcell.Color247,     // Rank
	tcell.Color247,     // File
	tcell.Color160,     // Prompt
	tcell.Color240,     // MeterBase
	tcell.ColorDefault, // MeterMid
	tcell.Color45,      // MeterNeutral
	tcell.Color122,     // MeterWin
	tcell.Color167,     // MeterLose
	tcell.ColorDefault, // PlayerNames
	tcell.Color247,     // Score
	tcell.ColorDefault, // MoveBox
	tcell.ColorDefault, // Emoji
	tcell.ColorDefault, // Input
	tcell.Color247,     // Advantage
}
