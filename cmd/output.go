package cmd

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/colors"
)

const (
	lineNbrIndentBy = 2
	lineSeparator   = "▐"
)

var lineNbrIndent = strings.Repeat(" ", lineNbrIndentBy)

func formatDimmedLine(maxLineDigits int, lineNbr int, lineContent string) string {
	return pp.Dimf("%s%*d %s %s\n", lineNbrIndent, maxLineDigits, lineNbr, lineSeparator, lineContent)
}

func formatElipsisLine(maxLineDigits int) string {
	return pp.Dimf("%s%*s %s\n", lineNbrIndent, maxLineDigits, "∙∙∙", lineSeparator)
}

func formatUncoveredLine(maxLineDigits int, lineNbr int, lineContent string) string {
	boldLineNbr := pp.Boldf("%*d", maxLineDigits, lineNbr)

	return fmt.Sprintf("%s%s%s\n", lineNbrIndent, boldLineNbr, pp.Redf(" %s %s", lineSeparator, lineContent))
}

// outputUncoveredLines writes the uncovered lines from the report to
// [os.Stdout]. If an output file path is specified, the uncovered lines will
// also be written to that file.
func outputUncoveredLines(report *analyzer.Report, outputFilePath string) error {
	if !report.HasUncoveredLines() {
		return nil
	}

	var outputFile *os.File

	if outputFilePath != "" {
		var err error
		outputFile, err = os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close() //nolint:errcheck
	}

	fmt.Printf("%s\n\n", colors.LightGreen("Missing coverage:"))

	for _, file := range report.Files {
		if len(file.NewUncoveredNewLinesGroups) == 0 {
			continue
		}

		// fmt.Println(pp.Bold(file.Path))

		for _, funcName := range file.ASTFile.FuncOrder {
			_, ok := file.NewUncoveredNewLinesGroups[funcName]
			if !ok {
				continue
			}

			outputUncoveredLineToStdout3(file, funcName)
		}

		// for funcName, lineRange := range file.NewUncoveredNewLinesGroups {
		// 	_ = funcName
		// 	_ = lineRange
		// outputUncoveredLineToStdout(file.Path, lineRange)
		// outputUncoveredLineToStdout3(file, funcName)
		// outputUncoveredLinetoFile(outputFile, file.Path, lineRange)
		// }

		// if len(file.NewUncoveredNewLinesGroups) > 0 {
		// 	fmt.Println()
		// }
	}

	return nil
}

// outputUncoveredLineToStdout writes the uncovered line range for a given file
// to the [os.Stdout].
func outputUncoveredLineToStdout(filePath string, lineRange analyzer.LineRange) {
	fmt.Printf("%s:%s:%s\n",
		pp.Bold(filePath),
		pp.Redf("%d", lineRange.Start),
		pp.Redf("%d", lineRange.End),
	)
}

// outputUncoveredLineToStdout writes the uncovered line range for a given file
// to the [os.Stdout].
func outputUncoveredLineToStdout3(file *analyzer.FileReport, funcName string) {
	f, ok := file.ASTFile.FuncDecls[funcName]
	if !ok {
		return
	}

	funcBodyLineStart := max(f.StartLine+1, 1)
	funcBodyLineEnd := min(f.EndLine-1, len(file.ASTFile.Lines))

	paddedUncoveredNewLines := padLines(file.UncoveredNewLines, funcBodyLineStart, funcBodyLineEnd)

	minLine := slices.Min(paddedUncoveredNewLines)
	maxLine := slices.Max(paddedUncoveredNewLines)
	maxLineDigits := max(len(strconv.Itoa(f.EndLine)), 3)

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s\n\n", pp.Boldf("%s:%d:%d", file.Path, f.StartLine, f.EndLine))
	fmt.Fprint(&buf, formatDimmedLine(maxLineDigits, f.StartLine, file.ASTFile.Lines[f.StartLine-1]))

	// If the first uncovered line is not immediately after the function start,
	// add an ellipsis line.
	if minLine > f.StartLine+1 {
		fmt.Fprint(&buf, formatElipsisLine(maxLineDigits))
	}

	for i, lineNbr := range paddedUncoveredNewLines {
		uncovered := slices.Contains(file.UncoveredNewLines, lineNbr)

		if !uncovered {
			fmt.Fprint(&buf, formatDimmedLine(maxLineDigits, lineNbr, file.ASTFile.Lines[lineNbr-1]))
		} else {
			fmt.Fprint(&buf, formatUncoveredLine(maxLineDigits, lineNbr, file.ASTFile.Lines[lineNbr-1]))
		}

		// If there is a gap between the current line and the next line, add an
		// ellipsis line.
		if i < len(paddedUncoveredNewLines)-1 {
			nextLineNbr := paddedUncoveredNewLines[i+1]
			if nextLineNbr > lineNbr+1 {
				fmt.Fprint(&buf, formatElipsisLine(maxLineDigits))
			}
		}
	}

	// If the last uncovered line is not immediately before the function end,
	// add an ellipsis line.
	if maxLine+1 < f.EndLine {
		fmt.Fprint(&buf, formatElipsisLine(maxLineDigits))
	}

	fmt.Fprint(&buf, formatDimmedLine(maxLineDigits, f.EndLine, file.ASTFile.Lines[f.EndLine-1]))
	fmt.Fprintln(&buf)

	fmt.Print(buf.String())
}

// outputUncoveredLinetoFile writes the uncovered line range for a given file to
// the specified output file.
func outputUncoveredLinetoFile(file *os.File, filePath string, lineRange analyzer.LineRange) {
	switch file {
	case nil:
		return
	default:
		_, _ = fmt.Fprintf(file, "%s:%d:%d\n", filePath, lineRange.Start, lineRange.End)
	}
}

func padLines(lines []int, funcBodyLineStart, funcBodyLineEnd int) []int {
	paddedUncoveredNewLines := make([]int, 0)
	for _, line := range lines {
		for i := line - 3; i <= line+3; i++ {
			if i >= funcBodyLineStart && i <= funcBodyLineEnd {
				paddedUncoveredNewLines = append(paddedUncoveredNewLines, i)
			}
		}
	}

	slices.Sort(paddedUncoveredNewLines)

	return slices.Compact(paddedUncoveredNewLines)
}
