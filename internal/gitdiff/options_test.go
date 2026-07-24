package gitdiff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [optionsToArgs] function.
func Test_optionsToArgs(t *testing.T) {
	t.Run("should return empty args for nil options", func(t *testing.T) {
		args := optionsToArgs(nil)
		require.Empty(t, args)
	})

	t.Run("should return empty args for empty options", func(t *testing.T) {
		args := optionsToArgs(&Options{})
		require.Len(t, args, 2)
		require.Equal(t, "... --unified=0", strings.Join(args, " "))
	})

	t.Run("should return args for unstaged changes", func(t *testing.T) {
		opts := &Options{
			Unstaged:  true,
			TargetRef: OriginMain,
		}
		args := optionsToArgs(opts)
		require.Len(t, args, 4)
		require.Equal(t, "origin/main -- . --unified=0", strings.Join(args, " "))
	})

	t.Run("should return args for staged changes", func(t *testing.T) {
		opts := &Options{
			TargetRef: OriginMain,
		}
		args := optionsToArgs(opts)
		require.Len(t, args, 2)
		require.Equal(t, "origin/main... --unified=0", strings.Join(args, " "))
	})
}
