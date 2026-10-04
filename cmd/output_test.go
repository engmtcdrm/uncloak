package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

const content = "test content"

// maxDigitsTestStruct is used for various format functions.
type maxDigitsTestStruct struct {
	maxLineDigits      int
	lineNbr            int
	lineContent        string
	expectedPaddingLen int
}

// Tests for [displayUncoveredFunctionLines] function.
func Test_displayUncoveredFunctionLines(t *testing.T) {
	const funcName = "testFunc"

	// newDisplayTestFile builds a [*analyzer.FileReport] with a single function
	// declaration spanning start to end lines and the given uncovered lines.
	newDisplayTestFile := func(t *testing.T, start, end int, uncoveredLines []int) *analyzer.FileReport {
		t.Helper()

		lines := make([]string, end+5)
		for i := range lines {
			lines[i] = fmt.Sprintf("line %d content", i+1)
		}

		return &analyzer.FileReport{
			Path:              "file.go",
			UncoveredNewLines: uncoveredLines,
			FuncUncoveredNewLines: map[string][]int{
				funcName: uncoveredLines,
			},
			ASTFile: &goast.File{
				Lines: lines,
				FuncDecls: goast.FuncDecls{
					funcName: {
						Name:  funcName,
						Start: start,
						End:   end,
						Body: goast.FuncBody{
							Start: start,
							End:   end,
						},
					},
				},
				FuncOrder: []string{funcName},
			},
		}
	}

	t.Run("should return early if the function is not found in the file", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(t, 1, 10, []int{5})
		displayUncoveredFunctionLines(file, "missingFunc")

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Empty(t, contents)
	})

	t.Run("should return early if the function has no uncovered lines entry", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(t, 1, 10, []int{5})
		file.FuncUncoveredNewLines = map[string][]int{}
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		require.Empty(t, contents)
	})

	t.Run("should not add ellipsis lines when uncovered lines are contiguous with the function boundaries", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(t, 1, 10, []int{2, 9})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "file.go:1:10")
		require.Zero(t, strings.Count(contentsNoANSI, ellipsis))
	})

	t.Run("should add ellipsis lines around the function boundaries and dim lines that are not uncovered", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(t, 1, 20, []int{10})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "line 7 content")
		require.Contains(t, contentsNoANSI, "line 10 content")
		require.Equal(t, 2, strings.Count(contentsNoANSI, ellipsis))
	})

	t.Run("should add an ellipsis line between groups of uncovered lines when there is a gap", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(t, 1, 22, []int{5, 21})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "line 5 content")
		require.Contains(t, contentsNoANSI, "line 21 content")
		require.Equal(t, 1, strings.Count(contentsNoANSI, ellipsis))
	})

	t.Run("should output single uncovered line when function is single line", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)

		file := newDisplayTestFile(t, 1, 1, []int{1})
		displayUncoveredFunctionLines(file, funcName)

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		contentsNoANSI := ansi.Strip(string(contents))

		require.Contains(t, contentsNoANSI, "line 1 content")
		require.Zero(t, strings.Count(contentsNoANSI, ellipsis))
	})
}

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

// Tests for [formatDimmedLine] function.
//
//nolint:dupl
func Test_formatDimmedLine(t *testing.T) {
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

// Tests for [formatEllipsisLine] function.
func Test_formatEllipsisLine(t *testing.T) {
	t.Run("should pad the line correctly", func(t *testing.T) {
		maxLineDigitsTests := []struct {
			maxLineDigits      int
			expectedPaddingLen int
		}{
			{1, 2},
			{2, 2},
			{3, 2},
			{4, 3},
			{5, 4},
		}

		for _, tt := range maxLineDigitsTests {
			t.Run(fmt.Sprintf("maxLineDigits=%d", tt.maxLineDigits), func(t *testing.T) {
				paddingLen := lineNbrIndentBy + max(tt.maxLineDigits-ellipsisLen, 0)
				require.Equal(t, tt.expectedPaddingLen, paddingLen)

				expectedPadding := strings.Repeat(" ", paddingLen)

				result := formatEllipsisLine(tt.maxLineDigits)
				resultNoANSI := strings.ReplaceAll(ansi.Strip(result), "\n", "")

				require.NotEmpty(t, result)
				require.Equal(t, expectedPadding, resultNoANSI[:paddingLen])
				require.Equal(t, fmt.Sprintf("%s %s", ellipsis, lineSeparator), resultNoANSI[paddingLen:])
			})
		}
	})
}

// Tests for [formatFuncSignature] function.
func Test_formatFuncSignature(t *testing.T) {
	tests := []struct {
		name          string
		maxLineDigits int
		start         int
		bodyStart     int
		lines         []string
		expected      []string
	}{
		{
			name:          "single-line signature",
			maxLineDigits: 1,
			start:         2,
			bodyStart:     2,
			lines:         []string{"package example", "func test() {", "}", ""},
			expected:      []string{"  2 ▐ func test() {"},
		},
		{
			name:          "multiline signature",
			maxLineDigits: 3,
			start:         3,
			bodyStart:     5,
			lines: []string{
				"package example",
				"",
				"func test(",
				"\tvalue int,",
				") {",
				"}",
			},
			expected: []string{
				"    3 ▐ func test(",
				"    4 ▐ \tvalue int,",
				"    5 ▐ ) {",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileReport := &analyzer.FileReport{
				ASTFile: &goast.File{Lines: tt.lines},
			}
			funcDecl := &goast.FuncDecl{
				Start: tt.start,
				Body:  goast.FuncBody{Start: tt.bodyStart},
			}

			result := formatFuncSignature(tt.maxLineDigits, fileReport, funcDecl)
			formattedLines := strings.Split(strings.TrimSuffix(ansi.Strip(result), "\n"), "\n")

			require.Equal(t, tt.expected, formattedLines)
		})
	}
}

// Tests for [formatLines] function.
func Test_formatLines(t *testing.T) {
	tests := []struct {
		name                    string
		lineNbr                 int
		uncoveredLines          []int
		paddedUncoveredNewLines []int
		idx                     int
		expected                string
	}{
		{
			name:                    "covered line without a gap",
			lineNbr:                 4,
			uncoveredLines:          []int{5},
			paddedUncoveredNewLines: []int{4, 5},
			idx:                     0,
			expected:                "   4 ▐ line 4 content\n",
		},
		{
			name:                    "uncovered line without a gap",
			lineNbr:                 4,
			uncoveredLines:          []int{4, 5},
			paddedUncoveredNewLines: []int{4, 5},
			idx:                     0,
			expected:                "   4 ▐ line 4 content\n",
		},
		{
			name:                    "gap before the next line",
			lineNbr:                 4,
			uncoveredLines:          []int{4, 7},
			paddedUncoveredNewLines: []int{4, 7},
			idx:                     0,
			expected:                "   4 ▐ line 4 content\n  ∙∙∙ ▐\n",
		},
		{
			name:                    "last line does not add an ellipsis",
			lineNbr:                 7,
			uncoveredLines:          []int{4, 7},
			paddedUncoveredNewLines: []int{4, 7},
			idx:                     1,
			expected:                "   7 ▐ line 7 content\n",
		},
	}

	lines := []string{
		"line 1 content",
		"line 2 content",
		"line 3 content",
		"line 4 content",
		"line 5 content",
		"line 6 content",
		"line 7 content",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileReport := &analyzer.FileReport{
				UncoveredNewLines: tt.uncoveredLines,
				ASTFile:           &goast.File{Lines: lines},
			}

			result := formatLines(fileReport, tt.lineNbr, 2, tt.paddedUncoveredNewLines, tt.idx)

			require.Equal(t, tt.expected, ansi.Strip(result))
		})
	}
}

// Tests for [formatUncoveredLine] function.
//
//nolint:dupl
func Test_formatUncoveredLine(t *testing.T) {
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

				result := formatUncoveredLine(tt.maxLineDigits, tt.lineNbr, tt.lineContent)
				resultNoANSI := strings.ReplaceAll(ansi.Strip(result), "\n", "")

				contentStartIdx := lineNbrIndentBy + tt.maxLineDigits + len(lineSeparator) + 2

				require.NotEmpty(t, result)
				require.Equal(t, expectedPadding, resultNoANSI[:paddingLen])
				require.Equal(t, tt.lineContent, resultNoANSI[contentStartIdx:])
			})
		}
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

	t.Run("should ignore if file has no uncovered lines", func(t *testing.T) {
		_, _, report := initReport(t)

		require.NotEmpty(t, report.Files)
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
	t.Run("should return early if file is nil", func(_ *testing.T) {
		outputUncoveredLinesToFile(nil, "file.go", []analyzer.LineRange{{Start: 1, End: 2}})
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

		uncoveredLines := []analyzer.LineRange{
			{Start: 1, End: 2},
			{Start: 5, End: 10},
		}

		outputUncoveredLinesToFile(file, "file.go", uncoveredLines)
		contents, err := os.ReadFile(tempFile)
		require.NoError(t, err)
		require.NotEmpty(t, contents)
		t.Logf("Uncovered lines written to file:\n%s", string(contents))

		for _, lineRange := range uncoveredLines {
			require.Contains(t, string(contents), fmt.Sprintf("file.go:%d:%d", lineRange.Start, lineRange.End))
		}
	})
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

	t.Run("should panic if lines is empty", func(t *testing.T) {
		lines := []int{}
		funcBodyLineStart := 1
		funcBodyLineEnd := 1

		padded := padLines(lines, funcBodyLineStart, funcBodyLineEnd)

		require.Panics(t, func() {
			_ = slices.Min(padded)
		})

		require.NotPanics(t, func() {
			if len(padded) > 0 {
				_ = slices.Min(padded)
			}
		})
	})

	t.Run("should panic if lines has values not between function body range", func(t *testing.T) {
		lines := []int{0, 8}
		funcBodyLineStart := 15
		funcBodyLineEnd := 20

		padded := padLines(lines, funcBodyLineStart, funcBodyLineEnd)

		require.Panics(t, func() {
			_ = slices.Min(padded)
		})

		require.NotPanics(t, func() {
			if len(padded) > 0 {
				_ = slices.Min(padded)
			}
		})
	})
}
