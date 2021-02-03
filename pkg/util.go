package uchess

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/notnil/chess"
)

// getSquare returns a chess square given a file and a rank
func getSquare(f chess.File, r chess.Rank) chess.Square {
	return chess.Square((int(r) * 8) + int(f))
}

// color calculates the color of the current square
func squareColor(sq chess.Square) chess.Color {
	if ((sq / 8) % 2) == (sq % 2) {
		return chess.Black
	}
	return chess.White
}

// IsInteractive returns a bool indicating whether the game is interactive
// with interactive meaning a human is involved
func IsInteractive(config Config) bool {
	return !(config.WhitePiece == "cpu" && config.BlackPiece == "cpu")
}

// IsCPU returns a bool indicating whether the specified player is the CPU
func IsCPU(player chess.Color, config Config) bool {
	if (player == chess.Black && config.BlackPiece == "cpu") ||
		(player == chess.White && config.WhitePiece == "cpu") {
		return true
	}
	return false
}

// WinProb calculates the percentage chance that white wins
func WinProb(cp int) float64 {
	pa := float64(cp) / 100
	return 1 / (1 + math.Pow(10, -pa/4.0))
}

// AtScale returns a boolean value indicating whether
// the given value (val) has crossed the threshold (idx)
// established by dividing total (tot) into equal
// increments
func AtScale(idx, tot int, val float64) bool {
	size := 1 / float64(tot) * 100
	if val >= float64(idx)*size {
		return true
	}
	return false
}

// Timestamp returns a timestamp string
func Timestamp() string {
	return time.Now().Format("20060102150405")
}

// InCheck returns two booleans indicating if either
// player is in check. The first boolean represents
// white and the second boolean represents black
func InCheck(game *chess.Game) (bool, bool) {
	playerTurn := game.Position().Turn()
	moves := game.Moves()
	if len(moves) == 0 {
		return false, false
	}
	lastMove := moves[len(moves)-1]
	inCheck := lastMove.HasTag(chess.Check)
	if inCheck && playerTurn == chess.White {
		return true, false
	} else if inCheck {
		return false, true
	}
	return false, false
}

func getCapturedPieces(pieces string, p, b, n, r, q, k string) string {
	pawns := 8 - strings.Count(pieces, p)
	bishops := 2 - strings.Count(pieces, b)
	knights := 2 - strings.Count(pieces, n)
	rooks := 2 - strings.Count(pieces, r)
	queens := 1 - strings.Count(pieces, q)
	kings := 1 - strings.Count(pieces, k)
	return strings.Repeat(p, pawns) +
		strings.Repeat(b, bishops) +
		strings.Repeat(n, knights) +
		strings.Repeat(r, rooks) +
		strings.Repeat(q, queens) +
		strings.Repeat(k, kings)
}

func getCaptured(pieces string) (string, string) {
	cWhite := getCapturedPieces(pieces, "P", "B", "N", "R", "Q", "K")
	cBlack := getCapturedPieces(pieces, "p", "b", "n", "r", "q", "k")
	return cWhite, cBlack
}

func getPiece(letter string) string {
	var m = map[string]string{
		"P": chess.WhitePawn.String(),
		"B": chess.WhiteBishop.String(),
		"N": chess.WhiteKnight.String(),
		"R": chess.WhiteRook.String(),
		"Q": chess.WhiteQueen.String(),
		"K": chess.WhiteKing.String(),
		"p": chess.WhitePawn.String(),
		"b": chess.WhiteBishop.String(),
		"n": chess.WhiteKnight.String(),
		"r": chess.WhiteRook.String(),
		"q": chess.WhiteQueen.String(),
		"k": chess.WhiteKing.String(),
	}
	return m[letter]
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func findStr(s []string, want string) int {
	for i, c := range s {
		if c == want {
			return i
		}
	}
	return -1
}

func disjoin(a, b string) string {
	result := strings.Split(strings.ToLower(a), "")
	for _, c := range b {
		idx := findStr(result, strings.ToLower(string(c)))
		if idx >= 0 {
			result = removeIndex(result, idx)
		}
	}
	return strings.Join(result, "")
}

func diffPieces(a, b string) (string, string) {
	return disjoin(b, a), disjoin(a, b)
}

// Advantages returns the white/black advantage strings
func Advantages(positions string) (string, string) {
	whiteCap, blackCap := getCaptured(positions)
	whiteAdv, blackAdv := diffPieces(whiteCap, blackCap)
	whiteRes := ""
	for _, c := range whiteAdv {
		whiteRes += getPiece(string(c))
	}
	blackRes := ""
	for _, c := range blackAdv {
		blackRes += getPiece(string(c))
	}
	return whiteRes, blackRes
}

// scorePieces returns the cumulative score for the piece string provided
// It expects a result from the Advantages function
func scorePieces(pieces string) int {
	score := 0
	var scores = map[rune]int{
		'r': 5,
		'n': 3,
		'b': 3,
		'q': 9,
		'k': 0,
		'p': 1,
	}
	for _, piece := range pieces {
		score += scores[piece]
	}
	return score
}

// ScoreStr returns a formatted score string
func ScoreStr(positions string) (string, string) {
	whiteCap, blackCap := getCaptured(positions)
	whiteAdv, blackAdv := diffPieces(whiteCap, blackCap)
	whiteScore := scorePieces(whiteAdv)
	blackScore := scorePieces(blackAdv)
	whiteDiff := whiteScore - blackScore
	blackDiff := blackScore - whiteScore
	whiteStr := ""
	blackStr := ""

	if whiteDiff > 0 {
		whiteStr = fmt.Sprintf("+%v", whiteDiff)
	}

	if blackDiff > 0 {
		blackStr = fmt.Sprintf("+%v", blackDiff)
	}
	return whiteStr, blackStr
}

// RoundNearest rounds to the nearest unit
func RoundNearest(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

// EmojiForPlayer returns an emoji corresponding to the player type
func EmojiForPlayer(playerType string) string {
	if playerType == "cpu" {
		return "🤖"
	}
	return "👤"
}

// IsWindows returns a boolean indicating whether windows is the OS
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// FileExists returns a bool indicating whether the specified file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsFile returns a bool indicating whether the specified file is a regular file
func IsFile(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return false
	}

	mode := fi.Mode()
	return mode.IsRegular()
}
