package gitdiff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [optionsToArgs] function.
func Test_optionsToArgs(t *testing.T) {
	t.Run("should return default args if options is empty", func(t *testing.T) {
		expectedArgs := []string{
			mergeBaseFlag,
			unifiedFlag,
			headRef,
			pathSpecSeparator,
		}
		expectedArgs = append(expectedArgs, goFileFilters...)

		args := optionsToArgs(Options{})

		require.Len(t, args, len(expectedArgs))
		require.Equal(t, strings.Join(expectedArgs, " "), strings.Join(args, " "))
	})

	t.Run("should return args with specified target-ref value", func(t *testing.T) {
		expectedArgs := []string{
			mergeBaseFlag,
			unifiedFlag,
			OriginMain,
			pathSpecSeparator,
		}

		opts := Options{
			TargetRef: OriginMain,
		}
		args := optionsToArgs(opts)

		expectedArgs = append(expectedArgs, goFileFilters...)
		require.Len(t, args, len(expectedArgs))
		require.Equal(t, strings.Join(expectedArgs, " "), strings.Join(args, " "))
	})
}
