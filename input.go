package uchess

import (
	"strings"
	"unicode/utf8"
)

// MaxLength defines the MaxWidth of the input buffer
const MaxLength = 10

// Input stores the input buffer
type Input struct {
	buffer string
}

// trimLastChar safely removes the last character in a string
func trimLastChar(s string) string {
	r, size := utf8.DecodeLastRuneInString(s)
	if r == utf8.RuneError && (size == 0 || size == 1) {
		size = 0
	}
	return s[:len(s)-size]
}

// NewInput creates a new input buffer
func NewInput() *Input {
	return &Input{""}
}

// Append appends to the input buffer
func (i *Input) Append(c rune) string {
	if len(i.buffer) < MaxLength {
		i.buffer += string(c)
	}
	return i.buffer
}

// Backspace removes the last entry in the input buffer
func (i *Input) Backspace() string {
	i.buffer = trimLastChar(i.buffer)
	return i.buffer
}

// Current returns the current input buffer
func (i *Input) Current() string {
	return i.buffer + strings.Repeat(" ", MaxLength-len(i.buffer))
}

// Length returns the input buffer length
func (i *Input) Length() int {
	return len(i.buffer)
}

// Clear clears the current buffer
func (i *Input) Clear() string {
	i.buffer = ""
	return i.buffer
}
