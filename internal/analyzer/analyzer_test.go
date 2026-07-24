package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [NewCodeCoverage] function.
func Test_NewCodeCoverage(t *testing.T) {
	cfg := &config.DefaultConfig
	cfg.GitDiffOptions.TargetRef = testgit.MainBranchName

	t.Run("should return a report without error", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		report, err := NewCodeCoverage(cfg)
		require.NoError(t, err)
		require.NotNil(t, report)
	})

	t.Run("should return an error if repository is not a git repository", func(t *testing.T) {
		t.Chdir(t.TempDir())

		report, err := NewCodeCoverage(cfg)
		require.Error(t, err)
		require.Empty(t, report)
	})

	t.Run("should return an error if there are no new lines from git diff", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		cfg := config.DefaultConfig
		cfg.GitDiffOptions.TargetRef = testrepo.NewBranchName

		report, err := NewCodeCoverage(&cfg)
		require.Error(t, err)
		require.Empty(t, report)
	})

	t.Run("should return an error when coverage is below threshold", func(t *testing.T) {
		tempDir, _ := testrepo.InitWithFileCopy(t)
		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		report, err := NewCodeCoverage(cfg)
		require.Error(t, err)
		require.NotNil(t, report)
	})
}

// Tests for [analyzeCoverage] function.
func Test_analyzeCoverage(t *testing.T) {
	cfg := &config.DefaultConfig
	cfg.GitDiffOptions.TargetRef = testgit.MainBranchName

	t.Run("should return a report with files", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		profile, diff, err := processFiles(cfg)
		require.NoError(t, err)

		report := NewReport(cfg.CoverageThreshold, profile, diff)
		require.NotNil(t, report)

		report = analyzeCoverage(report, cfg)
		require.NotNil(t, report)
		require.NotEmpty(t, report.GitDiffResults.Files())
	})

	t.Run("should return a report with excluded files", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		cfg := config.DefaultConfig
		cfg.Exclusions = []string{"utils/utils.go"}

		profile, diff, err := processFiles(&cfg)
		require.NoError(t, err)

		report := NewReport(cfg.CoverageThreshold, profile, diff)
		require.NotNil(t, report)

		report = analyzeCoverage(report, &cfg)
		require.NotNil(t, report)
		require.NotEmpty(t, report.GitDiffResults.Files())
	})
}

// Tests for [processFiles] function.
func Test_processFiles(t *testing.T) {
	cfg := &config.DefaultConfig
	cfg.GitDiffOptions.TargetRef = testgit.MainBranchName

	t.Run("should return a profile and diff without error", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		profile, diff, err := processFiles(cfg)
		require.NoError(t, err)
		require.NotNil(t, profile)
		require.NotNil(t, diff)
	})

	t.Run("should return error if repository is not a git repository", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		profile, diff, err := processFiles(cfg)
		require.Error(t, err)
		require.Nil(t, profile)
		require.Nil(t, diff)
	})
}
