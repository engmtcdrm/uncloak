package gocover

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Tests for [optionsToArgs] function.
func Test_optionsToArgs(t *testing.T) {
	t.Run("should return empty slice when options is nil", func(t *testing.T) {
		args := optionsToArgs(nil)
		require.Empty(t, args)
	})

	t.Run("should return empty slice when options has default values", func(t *testing.T) {
		opts := &Options{}
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})

	t.Run("should return slice with count argument when count is set", func(t *testing.T) {
		opts := &Options{Count: 3}
		expectedArgs := []string{"-count", "3"}

		args := optionsToArgs(opts)
		require.Equal(t, expectedArgs, args)
	})

	t.Run("should return empty slice if count is less than zero", func(t *testing.T) {
		opts := &Options{Count: -2}
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})

	t.Run("should return slice with timeout argument when timeout is set", func(t *testing.T) {
		opts := &Options{Timeout: 30 * time.Second}
		expectedArgs := []string{"-timeout", "30s"}

		args := optionsToArgs(opts)
		require.Equal(t, expectedArgs, args)
	})

	t.Run("should return empty slice if timeout is less than zero", func(t *testing.T) {
		opts := &Options{Timeout: -5 * time.Second}
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})

	t.Run("should return empty slice if verbose is false", func(t *testing.T) {
		opts := &Options{Verbose: false}
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})

	t.Run("should return slice with verbose argument when verbose is true", func(t *testing.T) {
		opts := &Options{Verbose: true}
		expectedArgs := []string{"-v"}

		args := optionsToArgs(opts)
		require.Equal(t, expectedArgs, args)
	})

	t.Run("should return slice when all options are set", func(t *testing.T) {
		opts := &Options{
			Count:   5,
			Timeout: 15 * time.Second,
			Verbose: true,
		}
		expectedArgs := []string{"-count", "5", "-timeout", "15s", "-v"}

		args := optionsToArgs(opts)
		require.Equal(t, expectedArgs, args)
	})
}
