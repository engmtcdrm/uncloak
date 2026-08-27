package gitdiff

import (
	"context"
	"strings"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [optionsToArgs] function.
func Test_optionsToArgs(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty args for nil options", func(t *testing.T) {
		args := optionsToArgs(ctx, nil)
		require.Empty(t, args)
	})

	t.Run("should return empty args for empty options", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		args := optionsToArgs(ctx, &Options{})
		require.Len(t, args, 4)
		require.Equal(t, "main -- . --unified=0", strings.Join(args, " "))
	})

	t.Run("should return args with specified target-ref value", func(t *testing.T) {
		opts := &Options{
			TargetRef: OriginMain,
		}
		args := optionsToArgs(ctx, opts)
		require.Len(t, args, 4)
		require.Equal(t, OriginMain+" -- . --unified=0", strings.Join(args, " "))
	})
}
