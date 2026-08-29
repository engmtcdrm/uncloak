package colors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [LightGreen] function.
func Test_LightGreen(t *testing.T) {
	text := "Hello, World!"
	coloredText := LightGreen(text)
	expected := "\x1b[38;2;169;209;61mHello, World!\x1b[0m"
	require.Equal(t, expected, coloredText)
}

// Tests for [Green] function.
func Test_Green(t *testing.T) {
	text := "Hello, World!"
	coloredText := Green(text)
	expected := "\x1b[38;2;97;179;79mHello, World!\x1b[0m"
	require.Equal(t, expected, coloredText)
}

// Tests for [Greenf] function.
func Test_Greenf(t *testing.T) {
	text := "Hello, World!"
	coloredText := Greenf("%s", text)
	expected := "\x1b[38;2;97;179;79mHello, World!\x1b[0m"
	require.Equal(t, expected, coloredText)
}

// Tests for [MediumGreen] function.
func Test_MediumGreen(t *testing.T) {
	text := "Hello, World!"
	coloredText := MediumGreen(text)
	expected := "\x1b[38;2;62;139;69mHello, World!\x1b[0m"
	require.Equal(t, expected, coloredText)
}

// Tests for [DarkGreen] function.
func Test_DarkGreen(t *testing.T) {
	text := "Hello, World!"
	coloredText := DarkGreen(text)
	expected := "\x1b[38;2;51;87;59mHello, World!\x1b[0m"
	require.Equal(t, expected, coloredText)
}
