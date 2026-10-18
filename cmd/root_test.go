package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [Execute] function.
func Test_Execute(t *testing.T) {
	ctx := context.Background()

	// Simple test helper to initialize a new root command and set required
	// flag(s). We want this due to the nature of the rootCmd being global
	// within this package. Otherwise flags could carry over.
	initValidCmd := func(t *testing.T) {
		t.Helper()

		rootCmd = newRootCmd()

		err := rootCmd.Flags().Set("target-ref", gitdiff.LocalMain)
		require.NoError(t, err)
	}

	t.Run("should run without error when in git repository", func(t *testing.T) {
		initValidCmd(t)
		_, _ = testrepo.InitWithFileCopy(ctx, t)

		err := Execute()
		require.NoError(t, err)
	})

	t.Run("should return error if coverage is below default", func(t *testing.T) {
		initValidCmd(t)
		tempDir, _ := testrepo.InitWithFileCopy(ctx, t)

		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		t.Chdir(tempDir)
		err = Execute()
		require.Error(t, err)

		coverageThresholdError := &coverageThresholdError{}
		require.ErrorAs(t, err, &coverageThresholdError)
	})

	t.Run("should return error if target-ref is not set", func(t *testing.T) {
		rootCmd = newRootCmd()
		err := Execute()
		require.Error(t, err)
		require.ErrorIs(t, err, errGitTargetRef)
	})
}
