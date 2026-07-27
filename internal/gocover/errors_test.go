package gocover

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [handleParseError] function.
func Test_handleParseError(t *testing.T) {
	t.Run("should return an error with command and output when exec fails with panic", func(t *testing.T) {
		cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
		err := handleParseError(cmd, []byte("output"), &exec.ExitError{Stderr: []byte("panic: something went wrong")})
		require.Error(t, err)
		require.Contains(t, err.Error(), "go test -coverprofile=coverage.out ./...")
		require.Contains(t, err.Error(), "output")
		require.Contains(t, err.Error(), "panic: something went wrong")
	})

	t.Run("should return an error with command and output when exec fails", func(t *testing.T) {
		cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
		err := handleParseError(cmd, []byte("output"), &exec.ExitError{Stderr: []byte("stderr")})
		require.Error(t, err)
		require.Contains(t, err.Error(), "go test -coverprofile=coverage.out ./...")
		require.Contains(t, err.Error(), "output")
		require.Contains(t, err.Error(), "stderr")
	})

	t.Run("should return an error with command and output when exec fails with unknown error", func(t *testing.T) {
		cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
		err := handleParseError(cmd, []byte("output"), errors.New("unknown error"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "go test -coverprofile=coverage.out ./...")
		require.Contains(t, err.Error(), "output")
	})
}
