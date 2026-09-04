package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
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

// Tests for [outputUncoveredLines] function.
func Test_outputUncoveredLines(t *testing.T) {
	initReport := func(t *testing.T) (tempDir string, stdoutFile *os.File, report *analyzer.Report) {
		t.Helper()

		ctx := context.Background()

		cfg := config.DefaultConfig
		cfg.GitDiffOptions.TargetRef = testgit.MainBranchName

		tempDir, stdoutFile = testrepo.InitWithFileCopy(ctx, t)
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

// Tests for [outputUncoveredLineToStdout] function.
func Test_outputUncoveredLineToStdout(t *testing.T) {
	t.Run("should output uncovered line to stdout", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		outputUncoveredLineToStdout("file.go", analyzer.LineRange{Start: 1, End: 2})

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.NotEmpty(t, contents)
		t.Logf("Uncovered lines written to stdout:\n%s", string(contents))
	})
}

// Tests for [outputUncoveredLinetoFile] function.
func Test_outputUncoveredLinetoFile(t *testing.T) {
	t.Run("should return early if file is nil", func(_ *testing.T) {
		outputUncoveredLinetoFile(nil, "file.go", analyzer.LineRange{Start: 1, End: 2})
	})

	t.Run("should write uncovered lines to file if valid", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "uncovered_lines.txt")
		file, err := os.Create(tempFile)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = file.Close()
			require.NoError(t, err)
		})

		outputUncoveredLinetoFile(file, "file.go", analyzer.LineRange{Start: 1, End: 2})
		contents, err := os.ReadFile(tempFile)
		require.NoError(t, err)
		require.NotEmpty(t, contents)
		t.Logf("Uncovered lines written to file:\n%s", string(contents))
	})
}
