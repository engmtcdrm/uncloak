package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/testing/testconfig"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [Execute] function.
func Test_Execute(t *testing.T) {
	t.Run("Execute runs without error when in git repository", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		require.NoError(t, Execute())
	})

	t.Run("Execute runs with error if coverage is below default", func(t *testing.T) {
		tempDir, _ := testrepo.InitWithFileCopy(t)
		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		t.Chdir(tempDir)
		require.Error(t, Execute())
	})
}

// Tests for [cmd.run] method.
func Test_run(t *testing.T) {
	t.Run("should run the command without error when in git repository", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		c := &cmd{}
		localRootCmd := rootCmd

		err := c.run(localRootCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("should run without error when in git repository and with arguments", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		c := &cmd{}
		localRootCmd := rootCmd

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

		err = c.run(localRootCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("should error when config file is invalid", func(t *testing.T) {
		tempDir, _ := testrepo.Init(t)
		c := &cmd{}
		localRootCmd := rootCmd

		testconfig.CreateConfigFile(t, tempDir, testconfig.InvalidUnknownFieldYaml)

		err := c.run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when NewCodeCoverage error is not a git repository", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := rootCmd
		_ = testutils.SetStdout(t)

		t.Chdir(t.TempDir())
		err := c.run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when coverage-threshold is negative", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		c := &cmd{}
		localRootCmd := rootCmd

		c.coverageThreshold = -1.0
		err := localRootCmd.Flags().Set("coverage-threshold", "-1.0")
		require.NoError(t, err)

		err = c.run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when output file cannot be written", func(t *testing.T) {
		tempDir, _ := testrepo.InitWithFileCopy(t)
		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		c := &cmd{}
		localRootCmd := rootCmd

		c.output = "/invalid/path/to/output/file"

		err = c.run(localRootCmd, []string{})
		require.Error(t, err)
	})
}

// Tests for [cmd.handleFlags] function.
func Test_cmd_handleFlags(t *testing.T) {
	t.Run("should set config values from flags", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		c := &cmd{}
		cfg := config.DefaultConfig
		localRootCmd := rootCmd

		c.coverageThreshold = 1.0
		err := localRootCmd.Flags().Set("coverage-threshold", "1.0")
		require.NoError(t, err)

		c.debug = true
		err = localRootCmd.Flags().Set("debug", "true")
		require.NoError(t, err)

		c.gitTargetRef = gitdiff.LocalMain
		err = localRootCmd.Flags().Set("target-ref", gitdiff.LocalMain)
		require.NoError(t, err)

		c.handleFlags(&cfg, localRootCmd)
		assert.Equal(t, 1.0, cfg.CoverageThreshold)
		assert.Equal(t, true, cfg.Debug)
		assert.Equal(t, gitdiff.LocalMain, cfg.GitDiffOptions.TargetRef)
	})
}

// Tests for [outputUncoveredLines] function.
func Test_outputUncoveredLines(t *testing.T) {
	initReport := func(t *testing.T) (tempDir string, stdoutFile *os.File, report *analyzer.Report) {
		t.Helper()

		cfg := config.DefaultConfig
		cfg.GitDiffOptions.TargetRef = testgit.MainBranchName

		tempDir, stdoutFile = testrepo.InitWithFileCopy(t)
		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		report, err = analyzer.NewCodeCoverage(&cfg)
		require.Error(t, err)
		require.NotNil(t, report)

		require.True(t, report.HasUncoveredLines())

		return tempDir, stdoutFile, report
	}

	t.Run("should return early if no uncovered lines exist", func(t *testing.T) {
		report := analyzer.NewReport(80.0, nil, nil)

		require.False(t, report.HasUncoveredLines())

		err := outputUncoveredLines(report, "")
		require.NoError(t, err)
	})

	t.Run("should output uncovered lines if they exist", func(t *testing.T) {
		_, _, report := initReport(t)

		err := outputUncoveredLines(report, "")
		require.NoError(t, err)
	})

	t.Run("should output to file if output path is specified", func(t *testing.T) {
		tempDir, _, report := initReport(t)

		outputFile := filepath.Join(tempDir, "uncovered_lines.txt")
		err := outputUncoveredLines(report, outputFile)
		require.NoError(t, err)

		contents, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		t.Logf("Uncovered lines written to %s:\n%s", outputFile, string(contents))
	})

	t.Run("should return error if output path is not writable", func(t *testing.T) {
		tempDir, _, report := initReport(t)

		outputFile := filepath.Join(tempDir, "non_existent_dir", "uncovered_lines.txt")
		err := outputUncoveredLines(report, outputFile)
		require.Error(t, err)
	})
}
