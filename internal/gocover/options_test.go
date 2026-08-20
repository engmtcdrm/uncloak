package gocover

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Tests for [optionsToArgs] function.
func Test_optionsToArgs(t *testing.T) {
	t.Run("should return empty slice when options is nil", func(t *testing.T) {
		var opts *Options = nil
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})

	t.Run("should return empty slice when options has default values", func(t *testing.T) {
		opts := &Options{}
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})

	t.Run("should return slice with timeout argument when timeout is set", func(t *testing.T) {
		opts := &Options{Timeout: 30 * time.Second}
		args := optionsToArgs(opts)
		require.Equal(t, []string{"--timeout", "30s"}, args)
	})

	t.Run("should return empty slice if timeout is less than zero", func(t *testing.T) {
		opts := &Options{Timeout: -5 * time.Second}
		args := optionsToArgs(opts)
		require.Empty(t, args)
	})
}
