package analyzer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
	"github.com/engmtcdrm/uncloak/internal/task"
	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [NewCodeCoverage] function.
func Test_NewCodeCoverage(t *testing.T) {
	cfg := config.DefaultConfig
	cfg.GitDiffOptions.TargetRef = testgit.MainBranchName
	testutils.SetStdout(t)

	t.Run("should return a report without error", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		report, err := NewCodeCoverage(&cfg)
		require.NoError(t, err)
		require.NotNil(t, report)
	})

	t.Run("should return a report without error when debug is true", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		_ = testutils.SetStdout(t)

		cfg := config.DefaultConfig
		cfg.Debug = true
		report, err := NewCodeCoverage(&cfg)
		require.NoError(t, err)
		require.NotNil(t, report)
	})

	t.Run("should return an error if repository is not a git repository", func(t *testing.T) {
		t.Chdir(t.TempDir())

		report, err := NewCodeCoverage(&cfg)
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

		report, err := NewCodeCoverage(&cfg)
		require.Error(t, err)
		require.NotNil(t, report)
	})

	t.Run("should return no error when none of the new lines are part of the coverage profile", func(t *testing.T) {
		repoPath := testgit.GetTestRepoPath(t)
		tempDir, _ := testrepo.Init(t)

		testfiles.CopyDir(t, repoPath, tempDir)
		testgit.AddCommit(t, "Add more files")

		readmePath := filepath.Join(tempDir, "README.md")
		testfiles.CreateFile(t, readmePath, "# Test README\nThis is a test README file.")
		testgit.CreateBranch(t, testrepo.NewBranchName)
		testgit.AddCommit(t, "Updated README.md")

		report, err := NewCodeCoverage(&cfg)
		require.NoError(t, err)
		require.NotNil(t, report)
	})
}

// Tests for [analyzeCoverage] function.
func Test_analyzeCoverage(t *testing.T) {
	cfg := config.DefaultConfig
	cfg.GitDiffOptions.TargetRef = testgit.MainBranchName
	testutils.SetStdout(t)

	t.Run("should return a report with files", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		profile, diff, err := processFiles(&cfg)
		require.NoError(t, err)

		report := NewReport(cfg.CoverageThreshold, profile, diff)
		require.NotNil(t, report)

		report = analyzeCoverage(report, &cfg)
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

// Tests for [filterFiles] function.
func Test_filterFiles(t *testing.T) {
	files := []string{"file1.go", "file2.go", "file3.go"}

	t.Run("should return all files if no exclusions are provided", func(t *testing.T) {
		cfg := config.DefaultConfig

		filteredFiles := filterFiles(&cfg, files)
		require.Equal(t, files, filteredFiles)
	})

	t.Run("should return filtered files if exclusions are provided", func(t *testing.T) {
		cfg := config.DefaultConfig
		cfg.Exclusions = []string{"file2.go"}
		expectedFiles := []string{"file1.go", "file3.go"}

		filteredFiles := filterFiles(&cfg, files)
		require.Equal(t, expectedFiles, filteredFiles)
	})
}

// Tests for [joinTaskErrors] function.
func Test_joinTaskErrors(t *testing.T) {
	t.Run("should return a nil error if no errors are provided", func(t *testing.T) {
		var errs []error

		joinedError := joinTaskErrors(errs...)
		require.NoError(t, joinedError)
	})

	t.Run("should return a single error if one error is provided", func(t *testing.T) {
		singleError := errors.New("single error")
		errs := []error{singleError}

		joinedError := joinTaskErrors(errs...)
		require.ErrorIs(t, joinedError, singleError)
	})

	t.Run("should return a combined error if multiple errors are provided", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		errs := []error{err1, err2}

		joinedError := joinTaskErrors(errs...)
		require.Error(t, joinedError)
		require.ErrorIs(t, joinedError, err1)
		require.ErrorIs(t, joinedError, err2)
	})

	t.Run("should not return errors that are taskCanceledError", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := taskCanceledError{}
		errs := []error{err1, err2}

		joinedError := joinTaskErrors(errs...)
		require.Error(t, joinedError)
		require.ErrorIs(t, joinedError, err1)
		require.NotErrorIs(t, joinedError, err2)
	})
}

// Tests for [printCommands] function.
func Test_printCommands(t *testing.T) {
	t.Run("should not print anything if both coverageProfile and diffResults are nil", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		printCommands(nil, nil)

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Empty(t, string(output))
	})

	t.Run("should print coverageProfile command if diffResults is nil", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		coverageProfile := gocover.NewProfile("go test -coverprofile=coverage.out", []byte{})
		printCommands(coverageProfile, nil)

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Contains(t, string(output), coverageProfile.Command)
	})

	t.Run("should print diffResults command if coverageProfile is nil", func(t *testing.T) {
		ctx := context.Background()
		stdoutFile := testutils.SetStdout(t)

		diffResults, err := gitdiff.NewResults(ctx, "git diff --name-only")
		require.NoError(t, err)

		printCommands(nil, diffResults)

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Contains(t, string(output), diffResults.Command)
	})

	t.Run("should print both coverageProfile and diffResults commands if both are provided", func(t *testing.T) {
		ctx := context.Background()
		stdoutFile := testutils.SetStdout(t)

		coverageProfile := gocover.NewProfile("go test -coverprofile=coverage.out", []byte{})
		diffResults, err := gitdiff.NewResults(ctx, "git diff --name-only")
		require.NoError(t, err)

		printCommands(coverageProfile, diffResults)

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Contains(t, string(output), coverageProfile.Command)
		require.Contains(t, string(output), diffResults.Command)
	})
}

// Tests for [processFiles] function.
func Test_processFiles(t *testing.T) {
	cfg := config.DefaultConfig
	cfg.GitDiffOptions.TargetRef = testgit.MainBranchName
	testutils.SetStdout(t)

	t.Run("should return a profile and diff without error", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		profile, diff, err := processFiles(&cfg)
		require.NoError(t, err)
		require.NotNil(t, profile)
		require.NotNil(t, diff)
	})

	t.Run("should return error if repository is not a git repository", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		profile, diff, err := processFiles(&cfg)
		require.Error(t, err)
		require.Nil(t, profile)
		require.Nil(t, diff)
	})
}

// Tests for [runTaskGitDiff] function.
func Test_runTaskGitDiff(t *testing.T) {
	ctx := context.Background()

	t.Run("should return diff results without error", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		stdoutFile := testutils.SetStdout(t)
		defaultConfig := config.DefaultConfig
		tm := task.NewManager()
		tm.Out = stdoutFile
		tm.Start()

		diffResults, err := runTaskGitDiff(ctx, tm, &defaultConfig.GitDiffOptions)
		require.NoError(t, err)
		require.NotNil(t, diffResults)
		require.NotEmpty(t, tm.Tasks)
		assert.Len(t, tm.Tasks, 1)

		tm.Finish()

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)

		outputStr := string(output)

		assert.Contains(t, outputStr, "✓", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✗", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "!", "Captured output:\n%s", outputStr)
	})

	t.Run("should return error if repository is not a git repository", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		stdoutFile := testutils.SetStdout(t)
		defaultConfig := config.DefaultConfig
		tm := task.NewManager()
		tm.Out = stdoutFile
		tm.Start()

		diffResults, err := runTaskGitDiff(ctx, tm, &defaultConfig.GitDiffOptions)
		require.Error(t, err)
		require.Nil(t, diffResults)
		require.NotEmpty(t, tm.Tasks)
		assert.Len(t, tm.Tasks, 1)

		tm.Finish()

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)

		outputStr := string(output)

		assert.Contains(t, outputStr, "✗", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✓", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "!", "Captured output:\n%s", outputStr)
	})

	t.Run("should return error with warning if context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately to simulate a canceled operation.

		_, _ = testrepo.InitWithFileCopy(t)

		stdoutFile := testutils.SetStdout(t)
		defaultConfig := config.DefaultConfig
		tm := task.NewManager()
		tm.Out = stdoutFile
		tm.Start()

		diffResults, err := runTaskGitDiff(ctx, tm, &defaultConfig.GitDiffOptions)
		require.Error(t, err)
		require.ErrorAs(t, err, &taskCanceledError{})
		require.ErrorAs(t, err, &context.Canceled)
		require.Nil(t, diffResults)
		require.NotEmpty(t, tm.Tasks)
		assert.Len(t, tm.Tasks, 1)

		tm.Finish()

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)

		outputStr := string(output)

		assert.Contains(t, outputStr, "!", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✓", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✗", "Captured output:\n%s", outputStr)
	})
}

// // Tests for [runTaskGoCoverage] function.
func Test_runTaskGoCoverage(t *testing.T) {
	ctx := context.Background()

	t.Run("should return coverage profile without error", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)

		stdoutFile := testutils.SetStdout(t)
		defaultConfig := config.DefaultConfig
		tm := task.NewManager()
		tm.Out = stdoutFile
		tm.Start()

		profile, err := runTaskGoCoverage(ctx, tm, &defaultConfig.GoTestOptions)
		require.NoError(t, err)
		require.NotNil(t, profile)
		require.NotEmpty(t, tm.Tasks)
		assert.Len(t, tm.Tasks, 1)

		tm.Finish()

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)

		outputStr := string(output)

		assert.Contains(t, outputStr, "✓", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✗", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "!", "Captured output:\n%s", outputStr)
	})

	t.Run("should return error if threshold is not met", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		stdoutFile := testutils.SetStdout(t)
		defaultConfig := config.DefaultConfig
		tm := task.NewManager()
		tm.Out = stdoutFile
		tm.Start()

		profile, err := runTaskGoCoverage(ctx, tm, &defaultConfig.GoTestOptions)
		require.Error(t, err)
		require.Nil(t, profile)
		require.NotEmpty(t, tm.Tasks)
		assert.Len(t, tm.Tasks, 1)

		tm.Finish()

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)

		outputStr := string(output)

		assert.Contains(t, outputStr, "✗", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✓", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "!", "Captured output:\n%s", outputStr)
	})

	t.Run("should return error with warning if context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately to simulate a canceled operation.

		_, _ = testrepo.InitWithFileCopy(t)

		stdoutFile := testutils.SetStdout(t)
		defaultConfig := config.DefaultConfig
		tm := task.NewManager()
		tm.Out = stdoutFile
		tm.Start()

		profile, err := runTaskGoCoverage(ctx, tm, &defaultConfig.GoTestOptions)
		require.Error(t, err)
		require.ErrorAs(t, err, &taskCanceledError{})
		require.ErrorAs(t, err, &context.Canceled)
		require.Nil(t, profile)
		require.NotEmpty(t, tm.Tasks)
		assert.Len(t, tm.Tasks, 1)

		tm.Finish()

		output, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)

		outputStr := string(output)

		assert.Contains(t, outputStr, "!", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✓", "Captured output:\n%s", outputStr)
		assert.NotContains(t, outputStr, "✗", "Captured output:\n%s", outputStr)
	})
}
