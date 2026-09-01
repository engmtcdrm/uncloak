package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/testing/testconfig"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [cmd.Run] function.
func Test_cmd_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("should run the command without error when in git repository", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		c := &cmd{}
		localRootCmd := newRootCmd()

		c.gitTargetRef = gitdiff.LocalMain
		err := localRootCmd.Flags().Set("target-ref", c.gitTargetRef)
		require.NoError(t, err)

		err = c.Run(localRootCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("should run without error when in git repository and with arguments", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		c := &cmd{}
		localRootCmd := newRootCmd()

		c.coverageThreshold = 1.0
		err := localRootCmd.Flags().Set("coverage-threshold", "1.0")
		require.NoError(t, err)

		c.debug = true
		err = localRootCmd.Flags().Set("debug", "true")
		require.NoError(t, err)

		c.gitTargetRef = gitdiff.LocalMain
		err = localRootCmd.Flags().Set("target-ref", gitdiff.LocalMain)
		require.NoError(t, err)

		c.verbose = true
		err = localRootCmd.Flags().Set("verbose", "true")
		require.NoError(t, err)

		err = c.Run(localRootCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("should error when config file is invalid", func(t *testing.T) {
		tempDir, _ := testrepo.Init(ctx, t)
		c := &cmd{}
		localRootCmd := newRootCmd()

		testconfig.CreateConfigFile(t, tempDir, testconfig.InvalidUnknownFieldYaml)

		err := c.Run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when NewCodeCoverage error is not a git repository", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := newRootCmd()
		_ = testutils.SetStdout(t)

		t.Chdir(t.TempDir())
		err := c.Run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when coverage-threshold is negative", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		c := &cmd{}
		localRootCmd := newRootCmd()

		c.coverageThreshold = -1.0
		err := localRootCmd.Flags().Set("coverage-threshold", "-1.0")
		require.NoError(t, err)

		err = c.Run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when output file cannot be written", func(t *testing.T) {
		tempDir, _ := testrepo.InitWithFileCopy(ctx, t)
		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		c := &cmd{}
		localRootCmd := newRootCmd()

		c.output = "/invalid/path/to/output/file"
		err = localRootCmd.Flags().Set("output", c.output)
		require.NoError(t, err)

		c.gitTargetRef = gitdiff.LocalMain
		err = localRootCmd.Flags().Set("target-ref", c.gitTargetRef)
		require.NoError(t, err)

		c.coverageThreshold = 25
		err = localRootCmd.Flags().Set("coverage-threshold", "25")
		require.NoError(t, err)

		err = c.Run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should run without returning an error if branches are the same", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		c := &cmd{}
		localRootCmd := newRootCmd()

		c.gitTargetRef = testrepo.NewBranchName
		err := localRootCmd.Flags().Set("target-ref", testrepo.NewBranchName)
		require.NoError(t, err)

		err = c.Run(localRootCmd, []string{})
		require.NoError(t, err)
	})
}

// Tests for [cmd.ValidateFlags] function.
func Test_cmd_ValidateFlags(t *testing.T) {
	t.Run("should return an error if target ref was not provided", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := &cobra.Command{}

		err := c.ValidateFlags(localRootCmd, nil)
		require.Error(t, err)
	})

	t.Run("should return nil if target ref is provided", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := newRootCmd()

		c.gitTargetRef = gitdiff.LocalMain
		err := localRootCmd.Flags().Set("target-ref", gitdiff.LocalMain)
		require.NoError(t, err)

		err = c.ValidateFlags(localRootCmd, nil)
		require.NoError(t, err)
	})

	t.Run("should return error if cmd.gitTargetRef is empty", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := newRootCmd()

		err := localRootCmd.Flags().Set("target-ref", "")
		require.NoError(t, err)

		err = c.ValidateFlags(localRootCmd, nil)
		require.Error(t, err)
	})

	t.Run("should return error if target ref flag is not set, but cmd.gitTargetRef is provided", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := newRootCmd()

		c.gitTargetRef = gitdiff.LocalMain

		err := c.ValidateFlags(localRootCmd, nil)
		require.Error(t, err)
	})
}

// Tests for [cmd.handleFlags] function.
func Test_cmd_handleFlags(t *testing.T) {
	ctx := context.Background()

	t.Run("should set config values from flags", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		c := &cmd{}
		cfg := config.DefaultConfig
		localRootCmd := newRootCmd()

		const expectedCoverageFile = "coverage.out"
		c.coverageFile = expectedCoverageFile
		err := localRootCmd.Flags().Set("coverage-file", expectedCoverageFile)
		require.NoError(t, err)

		c.coverageThreshold = 1.0
		err = localRootCmd.Flags().Set("coverage-threshold", "1.0")
		require.NoError(t, err)

		c.debug = true
		err = localRootCmd.Flags().Set("debug", "true")
		require.NoError(t, err)

		c.gitTargetRef = gitdiff.LocalMain
		err = localRootCmd.Flags().Set("target-ref", gitdiff.LocalMain)
		require.NoError(t, err)

		c.handleFlags(&cfg, localRootCmd)
		assert.Equal(t, expectedCoverageFile, cfg.CoverageFile)
		assert.InDelta(t, 1.0, cfg.CoverageThreshold, 0.1)
		assert.True(t, cfg.Debug)
		assert.Equal(t, gitdiff.LocalMain, cfg.GitDiffOptions.TargetRef)
	})
}
