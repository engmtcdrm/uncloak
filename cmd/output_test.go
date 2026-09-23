package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/engmtcdrm/go-ansi"
	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/require"
)

// Tests for [formatDimmedLine] function.
func Test_formatDimmedLine(t *testing.T) {
	const content = "test content"

	type maxDigitsTestStruct struct {
		maxLineDigits      int
		lineNbr            int
		lineContent        string
		expectedPaddingLen int
	}

	t.Run("should pad the line correctly", func(t *testing.T) {
		maxLineDigitsTests := []maxDigitsTestStruct{
			{3, 1, content, 4},
			{3, 10, content, 3},
			{3, 100, content, 2},
			{4, 1, content, 5},
			{4, 10, content, 4},
			{4, 100, content, 3},
			{4, 1000, content, 2},
		}

		for _, tt := range maxLineDigitsTests {
			t.Run(fmt.Sprintf("maxLineDigits=%d,lineNbr=%d", tt.maxLineDigits, tt.lineNbr), func(t *testing.T) {
				lineNbrLen := len(strconv.Itoa(tt.lineNbr))
				paddingLen := lineNbrIndentBy + tt.maxLineDigits - lineNbrLen
				require.Equal(t, tt.expectedPaddingLen, paddingLen)

				expectedPadding := strings.Repeat(" ", paddingLen)

				result := formatDimmedLine(tt.maxLineDigits, tt.lineNbr, tt.lineContent)
				resultNoANSI := strings.ReplaceAll(ansi.Strip(result), "\n", "")

				contentStartIdx := lineNbrIndentBy + tt.maxLineDigits + len(lineSeparator) + 2

				require.NotEmpty(t, result)
				require.Equal(t, expectedPadding, resultNoANSI[:paddingLen])
				require.Equal(t, tt.lineContent, resultNoANSI[contentStartIdx:])
			})
		}
	})
}

// Tests for [formatElipsisLine] function.
func Test_formatElipsisLine(t *testing.T) {
	t.Run("ss", func(t *testing.T) {
		result := formatElipsisLine(3)
		result += formatElipsisLine(4)
		result += formatDimmedLine(5, 1001, "some more content")
		result += formatElipsisLine(5)
		result += formatDimmedLine(5, 10001, "some more content")
		require.NotEmpty(t, result)
	})
}

// Tests for [formatUncoveredLine] function.
func Test_formatUncoveredLine(t *testing.T) {
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

// Tests for [displayUncoveredLines] function.
func Test_displayUncoveredLines(t *testing.T) {
	t.Run("should output uncovered line to stdout", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		displayUncoveredLines("file.go", analyzer.LineRange{Start: 1, End: 2})

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.NotEmpty(t, contents)
		t.Logf("Uncovered lines written to stdout:\n%s", string(contents))
	})
}

// Tests for [outputUncoveredLineToFile] function.
func Test_outputUncoveredLineToFile(t *testing.T) {
	t.Run("should return early if file is nil", func(_ *testing.T) {
		outputUncoveredLineToFile(nil, "file.go", analyzer.LineRange{Start: 1, End: 2})
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

		outputUncoveredLineToFile(file, "file.go", analyzer.LineRange{Start: 1, End: 2})
		contents, err := os.ReadFile(tempFile)
		require.NoError(t, err)
		require.NotEmpty(t, contents)
		t.Logf("Uncovered lines written to file:\n%s", string(contents))
	})
}

// Tests for [outputUncoveredLinesToFile] function.
func Test_outputUncoveredLinesToFile(t *testing.T) {
}
