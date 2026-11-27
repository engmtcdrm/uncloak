package gocover

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/cover"
)

const doesNotExist = "does-not-exist"

// Tests for [Run] function.
func Test_Run(t *testing.T) {
	opts := &DefaultOptions

	t.Run("should return error if go list fails", func(t *testing.T) {
		t.Chdir(t.TempDir())

		_, err := Run(t.Context(), "", opts)
		require.Error(t, err)
	})

	t.Run("should return valid profile for empty filePath", func(t *testing.T) {
		tempDir := t.TempDir()
		repoPath := testgit.GetTestRepoPath(t.Context(), t)
		testfiles.CopyDir(t, repoPath, tempDir)
		t.Chdir(tempDir)

		profile, err := Run(t.Context(), "", opts)
		require.NoError(t, err)
		require.NotNil(t, profile)
	})

	t.Run("should return valid profile for nil options", func(t *testing.T) {
		tempDir := t.TempDir()
		repoPath := testgit.GetTestRepoPath(t.Context(), t)
		testfiles.CopyDir(t, repoPath, tempDir)
		t.Chdir(tempDir)

		profile, err := Run(t.Context(), "", nil)
		require.NoError(t, err)
		require.NotNil(t, profile)
	})

	t.Run("should return error if temp directory does not exist", func(t *testing.T) {
		tempDir := t.TempDir() + doesNotExist

		// On Linux and other Unix-like systems, TMPDIR is used over /tmp, so we
		// set it to a directory that does not exist. On Windows we set TMP to a
		// directory that does not exist.
		switch runtime.GOOS {
		case "windows":
			t.Setenv("TMP", tempDir)
		default:
			t.Setenv("TMPDIR", tempDir)
		}

		profile, err := Run(t.Context(), "", opts)
		require.Error(t, err)
		require.NotNil(t, profile)
		require.Empty(t, profile.Command)
		require.Empty(t, profile.RawTestOutput)
	})
}

// Tests for [parser.parseProfiles] function.
func Test_parser_parseProfiles(t *testing.T) {
	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error if file does not exist", func(t *testing.T) {
		results, err := p.parseProfiles("nonexistent_file.txt")
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("should return error if file cannot be read", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping test on Windows as opening a directory may not fail with EISDIR")
		}

		// Opening a directory succeeds, but reading from it fails with EISDIR.
		results, err := p.parseProfiles(t.TempDir())
		require.Error(t, err)
		require.Nil(t, results)
		assert.ErrorIs(t, err, ErrFailedToReadFile)
	})

	t.Run("should return error when coverage file is empty", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "coverage.out")

		tmpFile, err := os.Create(tempFile)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		results, err := p.parseProfiles(tempFile)
		require.Error(t, err)
		require.Nil(t, results)
		assert.ErrorIs(t, err, ErrEmptyCoverageFile)
	})

	t.Run("should return no error if no lines to test", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "coverage.out")
		coverageData := "mode: set"

		testfiles.CreateFile(t, tempFile, coverageData)

		results, err := p.parseProfiles(tempFile)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Empty(t, results.CoveredLines, 0)
		require.Empty(t, results.Lines, 0)
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

		expectedProfiles := NewProfile("", []byte{})
		expectedProfiles.Lines = []Line{
			{
				FilePath:    "file1.go",
				Path:        "file1.go",
				Mode:        ModeSet,
				StartLine:   10,
				StartColumn: 1,
				EndLine:     20,
				EndColumn:   2,
				Statements:  1,
				Count:       1,
			},
			{
				FilePath:    "file2.go",
				Path:        "file2.go",
				Mode:        ModeSet,
				StartLine:   30,
				StartColumn: 3,
				EndLine:     40,
				EndColumn:   4,
				Statements:  2,
				Count:       0,
			},
		}

		for _, block := range expectedProfiles.Lines {
			expectedProfiles.expandCoveredLines(block)
		}

		profiles, err := p.parseProfiles(tempFile)
		require.NoError(t, err)
		require.NotNil(t, profiles)
		assert.Len(t, profiles.CoveredLines, 2)
		assert.Equal(t, expectedProfiles, profiles)
	})

	t.Run("should return error if block format is invalid", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "coverage.out")
		coverageData := `mode: set
invalid block format
`
		testfiles.CreateFile(t, tempFile, coverageData)

		profile, err := p.parseProfiles(tempFile)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidCoverageFile)
		require.Nil(t, profile)
	})

	t.Run("should return error if mode is missing", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "coverage.out")
		coverageData := `mode:
test.com/testmodule/file1.go:10.1,20.2 1 1
`
		testfiles.CreateFile(t, tempFile, coverageData)

		profiles, err := p.parseProfiles(tempFile)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidCoverageFile)
		require.Nil(t, profiles)
	})
}

// Tests for [parser.parseProfilesBlocks] function.
func Test_parser_parseProfilesBlocks(t *testing.T) {
	const (
		modulePath = "test.com/testmodule"

		file1           = "file1.go"
		file2           = "file2.go"
		file1ModulePath = modulePath + "/" + file1
		file2ModulePath = modulePath + "/" + file2

		validMode   = "set"
		invalidMode = "invalid"
	)

	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error for if profiles is nil", func(t *testing.T) {
		results, err := p.parseProfilesBlocks(nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrEmptyCoverageFile)
		require.Nil(t, results)
	})

	t.Run("should return valid profile for valid input", func(t *testing.T) {
		profiles := []*cover.Profile{
			{
				FileName: file1ModulePath,
				Mode:     validMode,
				Blocks: []cover.ProfileBlock{
					{
						StartLine: 10,
						StartCol:  1,
						EndLine:   20,
						EndCol:    2,
						NumStmt:   1,
						Count:     1,
					},
				},
			},
			{
				FileName: file2ModulePath,
				Mode:     validMode,
				Blocks: []cover.ProfileBlock{
					{
						StartLine: 30,
						StartCol:  3,
						EndLine:   40,
						EndCol:    4,
						NumStmt:   2,
						Count:     0,
					},
				},
			},
		}

		results, err := p.parseProfilesBlocks(profiles)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.CoveredLines, 2)
		assert.Len(t, results.Lines, 2)
	})

	t.Run("should return no error if no lines to test", func(t *testing.T) {
		profiles := []*cover.Profile{}

		results, err := p.parseProfilesBlocks(profiles)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Empty(t, results.CoveredLines, 0)
		assert.Empty(t, results.Lines, 0)
	})

	t.Run("should treat line as covered if any block touching it was hit", func(t *testing.T) {
		profiles := []*cover.Profile{
			{
				FileName: file1ModulePath,
				Mode:     validMode,
				Blocks: []cover.ProfileBlock{
					{
						StartLine: 10,
						StartCol:  1,
						EndLine:   11,
						EndCol:    2,
						NumStmt:   1,
						Count:     1,
					},
					{
						StartLine: 11,
						StartCol:  5,
						EndLine:   12,
						EndCol:    2,
						NumStmt:   1,
						Count:     0,
					},
				},
			},
		}

		expectedCoveredLines := map[int]bool{
			10: true,
			11: true,
			12: false,
		}

		results, err := p.parseProfilesBlocks(profiles)
		require.NoError(t, err)

		coveredLines, ok := results.CoveredLines[file1]
		require.True(t, ok)
		require.Equal(t, expectedCoveredLines, coveredLines)
	})

	t.Run("should merge identical blocks from multiple packages", func(t *testing.T) {
		profiles := []*cover.Profile{
			{
				FileName: file1ModulePath,
				Mode:     validMode,
				Blocks: []cover.ProfileBlock{
					{
						StartLine: 10,
						StartCol:  1,
						EndLine:   11,
						EndCol:    2,
						NumStmt:   1,
						Count:     0,
					},
					{
						StartLine: 10,
						StartCol:  1,
						EndLine:   11,
						EndCol:    2,
						NumStmt:   1,
						Count:     1,
					},
				},
			},
		}

		results, err := p.parseProfilesBlocks(profiles)
		require.NoError(t, err)

		coveredLines, ok := results.CoveredLines[file1]
		require.True(t, ok)
		assert.Contains(t, coveredLines, 10)
		assert.True(t, results.IsInTestCoverage(file1, 10))
	})

	t.Run("should return error if mode is invalid", func(t *testing.T) {
		profiles := []*cover.Profile{
			{
				FileName: file1ModulePath,
				Mode:     invalidMode,
				Blocks: []cover.ProfileBlock{
					{
						StartLine: 10,
						StartCol:  1,
						EndLine:   20,
						EndCol:    2,
						NumStmt:   1,
						Count:     1,
					},
				},
			},
		}

		results, err := p.parseProfilesBlocks(profiles)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidCoverageFile)
		require.ErrorIs(t, err, ErrInvalidMode)
		require.Nil(t, results)
	})
}

// Tests for [parser.runTestCoverage] function.
func Test_parser_runTestCoverage(t *testing.T) {
	p := &parser{
		GoList: &GoList{
			Module: "test.com/testmodule",
		},
	}

	t.Run("should return error if temp directory does not exist", func(t *testing.T) {
		tempDir := t.TempDir() + doesNotExist

		// On Linux and other Unix-like systems, TMPDIR is used over /tmp, so we
		// set it to a directory that does not exist. On Windows we set TMP to a
		// directory that does not exist.
		switch runtime.GOOS {
		case "windows":
			t.Setenv("TMP", tempDir)
		default:
			t.Setenv("TMPDIR", tempDir)
		}

		filePath, err := p.runTestCoverage(t.Context())
		t.Cleanup(func() {
			os.RemoveAll(tempDir) //nolint:errcheck
		})

		require.Error(t, err)
		require.Empty(t, filePath)
	})

	t.Run("should return error if go test fails", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		filePath, err := p.runTestCoverage(t.Context())
		t.Cleanup(func() {
			os.RemoveAll(tempDir) //nolint:errcheck
		})

		require.Error(t, err)
		require.Empty(t, filePath)
	})

	t.Run("should return valid file path for successful test coverage run", func(t *testing.T) {
		tempDir := t.TempDir()
		repoPath := testgit.GetTestRepoPath(t.Context(), t)
		testfiles.CopyDir(t, repoPath, tempDir)

		t.Chdir(tempDir)
		filePath, err := p.runTestCoverage(t.Context())
		t.Cleanup(func() {
			os.RemoveAll(tempDir) //nolint:errcheck
		})

		require.NoError(t, err)
		require.NotEmpty(t, filePath)
	})
}
