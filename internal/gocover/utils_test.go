package gocover

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [parseMode] function.
func Test_parseMode(t *testing.T) {
	t.Run("should return the correct mode", func(t *testing.T) {
		expectedMode := ModeSet
		result, err := parseMode("set")
		require.NoError(t, err)
		require.Equal(t, expectedMode, result)
	})

	t.Run("should return an error for an invalid mode", func(t *testing.T) {
		expectedMode := Mode("")
		result, err := parseMode("invalid")
		require.Error(t, err)
		require.Equal(t, expectedMode, result)
	})
}
