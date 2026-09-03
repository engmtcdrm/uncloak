package gocover

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/require"
)

// Tests for [Run] function.
func Test_Run(t *testing.T) {
	ctx := context.Background()
	opts := &DefaultOptions

	t.Run("should return error if go list fails", func(t *testing.T) {
		t.Chdir(t.TempDir())

		_, err := Run(ctx, "", opts)
		require.Error(t, err)
	})

	t.Run("should return valid profile for empty filePath", func(t *testing.T) {
		tempDir := t.TempDir()
		repoPath := testgit.GetTestRepoPath(ctx, t)
		testfiles.CopyDir(t, repoPath, tempDir)
		t.Chdir(tempDir)

		profile, err := Run(ctx, "", opts)
		require.NoError(t, err)
		require.NotNil(t, profile)
	})

	t.Run("should return valid profile for nil options", func(t *testing.T) {
		tempDir := t.TempDir()
		repoPath := testgit.GetTestRepoPath(ctx, t)
		testfiles.CopyDir(t, repoPath, tempDir)
		t.Chdir(tempDir)

		profile, err := Run(ctx, "", nil)
		require.NoError(t, err)
		require.NotNil(t, profile)
	})

	t.Run("should return error if temp directory cannot be written to", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping test on Windows due to permission issues with temp directories.")
		}

		tempDir := t.TempDir()
		// On linux TMPDIR is used over /tmp, so we set TMPDIR to a directory
		// with no write permissions
		t.Setenv("TMPDIR", tempDir)
		err := os.Chmod(tempDir, 0000)
		require.NoError(t, err)

		profile, err := Run(ctx, "", opts)
		require.Error(t, err)
		require.NotNil(t, profile)
		require.Empty(t, profile.Command)
		require.Empty(t, profile.RawTestOutput)
	})
}

// Tests for [parser.parseCoverageData] function.
func Test_parser_parseCoverageData(t *testing.T) {
	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error if scanner encounters an error", func(t *testing.T) {
		r := &testutils.ErrorReader{}
		results, err := p.parseCoverageData(r)
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("should return nil results for empty input", func(t *testing.T) {
		r := &testutils.EmptyReader{}
		results, err := p.parseCoverageData(r)
		require.NoError(t, err)
		require.Nil(t, results)
	})

	t.Run("should return valid profile for valid input", func(t *testing.T) {
		input := `mode: set
test.com/testmodule/file1.go:10.1,20.2 1 1
test.com/testmodule/file2.go:30.3,40.4 2 0
`
		r := bytes.NewBufferString(input)
		results, err := p.parseCoverageData(r)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, ModeSet, results.Mode)
		require.Len(t, results.CoveredLines, 2)
	})
}

// Tests for [parser.parseCoverageProfile] function.
func Test_parser_parseCoverageProfile(t *testing.T) {
	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error if file does not exist", func(t *testing.T) {
		results, err := p.parseCoverageProfile("nonexistent_file.txt")
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("should return valid profile for valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "coverage.out")
		// Create a temporary file with valid coverage data
		coverageData := `mode: set
test.com/testmodule/file1.go:10.1,20.2 1 1
test.com/testmodule/file2.go:30.3,40.4 2 0
`
		testfiles.CreateFile(t, tempFile, coverageData)

		results, err := p.parseCoverageProfile(tempFile)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, ModeSet, results.Mode)
		require.Len(t, results.CoveredLines, 2)
	})
}

// Tests for [parser.parseLines] function.
func Test_parser_parseLines(t *testing.T) {
	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error if line length is less than 2", func(t *testing.T) {
		profile, err := p.parseLines([]string{"mode: set"})
		require.Error(t, err)
		require.Nil(t, profile)
	})

	t.Run("should return error if lines are invalid", func(t *testing.T) {
		lines := []string{
			"mode: set",
			"invalid line format",
		}

		profile, err := p.parseLines(lines)
		require.Error(t, err)
		require.Nil(t, profile)
	})

	t.Run("should return error if Sscanf fails", func(t *testing.T) {
		lines := []string{
			"mode: set",
			"test.com/testmodule/file1.go:10.1,20.2 1 notanumber",
		}

		profile, err := p.parseLines(lines)
		require.Error(t, err)
		require.Nil(t, profile)
	})

	t.Run("should return valid profile if lines are valid", func(t *testing.T) {
		lines := []string{
			"mode: set",
			"test.com/testmodule/file1.go:10.1,20.2 1 1",
			"test.com/testmodule/file2.go:30.3,40.4 2 0",
		}

		profile, err := p.parseLines(lines)
		require.NoError(t, err)
		require.NotNil(t, profile)
		require.Equal(t, ModeSet, profile.Mode)
		require.Len(t, profile.CoveredLines, 2)
	})
}

// Tests for [parser.runTestCoverage] function.
func Test_parser_runTestCoverage(t *testing.T) {
	ctx := context.Background()
	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error if temp directory cannot be written to", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping test on Windows due to permission issues with temp directories.")
		}

		tempDir := t.TempDir()
		// On linux TMPDIR is used over /tmp, so we set TMPDIR to a directory
		// with no write permissions
		t.Setenv("TMPDIR", tempDir)
		err := os.Chmod(tempDir, 0000)
		require.NoError(t, err)

		filePath, err := p.runTestCoverage(ctx)
		require.Error(t, err)
		require.Empty(t, filePath)
	})

	t.Run("should return error if go test fails", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		filePath, err := p.runTestCoverage(ctx)
		require.Error(t, err)
		require.Empty(t, filePath)
	})

	t.Run("should return valid file path for successful test coverage run", func(t *testing.T) {
		tempDir := t.TempDir()
		repoPath := testgit.GetTestRepoPath(ctx, t)
		testfiles.CopyDir(t, repoPath, tempDir)

		t.Chdir(tempDir)
		filePath, err := p.runTestCoverage(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, filePath)
	})
}
