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
	"github.com/engmtcdrm/uncloak/internal/goast"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/require"
)

// Tests for [displayUncoveredLine] function.
func Test_displayUncoveredLine(t *testing.T) {
	t.Run("should output uncovered line to stdout", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		displayUncoveredLine("file.go", analyzer.LineRange{Start: 1, End: 2})

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "file.go:1:2")

		t.Logf("Uncovered lines written to stdout:\n%s", string(contents))
	})
}

// Tests for [displayUncoveredLines] function.
func Test_displayUncoveredLines(t *testing.T) {
	t.Run("should output uncovered lines to stdout", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		displayUncoveredLines("file.go", []analyzer.LineRange{
			{Start: 1, End: 2},
			{Start: 3, End: 4},
		})

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "file.go:1:2")
		require.Contains(t, contentsNoANSI, "file.go:3:4")
		t.Logf("Uncovered lines written to stdout:\n%s", string(contents))
	})
}

// Tests for [displayUncoveredFunctionLines] function.
func Test_displayUncoveredFunctionLines(t *testing.T) {
	const funcName = "testFunc"

	// newDisplayTestFile builds a [*analyzer.FileReport] with a single function
	// declaration spanning startLine to endLine and the given uncovered lines.
	newDisplayTestFile := func(startLine, endLine int, uncoveredLines []int) *analyzer.FileReport {
		lines := make([]string, endLine+5)
		for i := range lines {
			lines[i] = fmt.Sprintf("line %d content", i+1)
		}

		return &analyzer.FileReport{
			Path:              "file.go",
			UncoveredNewLines: uncoveredLines,
			ASTFile: &goast.File{
				Lines: lines,
				FuncDecls: goast.FuncDecls{
					funcName: {
						Name:      funcName,
						StartLine: startLine,
						EndLine:   endLine,
					},
				},
				FuncOrder: []string{funcName},
			},
		}
	}

	t.Run("should return early if the function is not found in the file", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(1, 10, []int{5})
		displayUncoveredFunctionLines(file, "missingFunc")

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Empty(t, contents)
	})

	t.Run("should not add ellipsis lines when uncovered lines are contiguous with the function boundaries", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(1, 10, []int{2, 9})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "file.go:1:10")
		require.Zero(t, strings.Count(contentsNoANSI, "∙∙∙"))
	})

	t.Run("should add ellipsis lines around the function boundaries and dim lines that are not uncovered", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(1, 20, []int{10})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "line 7 content")
		require.Contains(t, contentsNoANSI, "line 10 content")
		require.Equal(t, 2, strings.Count(contentsNoANSI, "∙∙∙"))
	})

	t.Run("should add an ellipsis line between groups of uncovered lines when there is a gap", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(1, 22, []int{5, 21})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "line 5 content")
		require.Contains(t, contentsNoANSI, "line 21 content")
		require.Equal(t, 1, strings.Count(contentsNoANSI, "∙∙∙"))
	})
}

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

	t.Run("should ignore if file has no uncovered lines", func(t *testing.T) {
		_, _, report := initReport(t)

		require.Greater(t, len(report.Files), 0)
		report.Files[0].FuncUncoveredNewLinesGroups = nil
		err := outputUncoveredLines(report, "")
		require.NoError(t, err)
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

// Tests for [padLines] function.
func Test_padLines(t *testing.T) {
	t.Run("should pad lines within function body range", func(t *testing.T) {
		lines := []int{3, 5}
		funcBodyLineStart := 1
		funcBodyLineEnd := 7

		padded := padLines(lines, funcBodyLineStart, funcBodyLineEnd)
		require.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, padded)
	})

	t.Run("should not pad lines outside function body range", func(t *testing.T) {
		lines := []int{1, 7}
		funcBodyLineStart := 2
		funcBodyLineEnd := 6

		padded := padLines(lines, funcBodyLineStart, funcBodyLineEnd)
		require.Equal(t, []int{2, 3, 4, 5, 6}, padded)
	})
}
