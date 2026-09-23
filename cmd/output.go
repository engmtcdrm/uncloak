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
	lineNbrIndentBy = 2   // Number of spaces to indent the line number column
	lineSeparator   = "▐" // Separator between the line number and the content
	padLinesBy      = 3   // Number of lines to pad before and after each uncovered line
)

var lineNbrIndent = strings.Repeat(" ", lineNbrIndentBy)

// displayUncoveredLines displays the uncovered line range for a given file to
// [os.Stdout].
func displayUncoveredLines(filePath string, lineRange analyzer.LineRange) {
	fmt.Printf("%s:%s:%s\n",
		pp.Bold(filePath),
		pp.Redf("%d", lineRange.Start),
		pp.Redf("%d", lineRange.End),
	)
}

// displayUncoveredFunctionLines displays the uncovered lines for a given
// function within a file to [os.Stdout].
func displayUncoveredFunctionLines(file *analyzer.FileReport, funcName string) {
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

// formatDimmedLine formats a line with dimmed text for the line number and
// content.
func formatDimmedLine(maxLineDigits int, lineNbr int, lineContent string) string {
	return pp.Dimf("%s%*d %s %s\n", lineNbrIndent, maxLineDigits, lineNbr, lineSeparator, lineContent)
}

// formatElipsisLine formats an ellipsis line with dimmed text for the line
// number column.
func formatElipsisLine(maxLineDigits int) string {
	return pp.Dimf("%s%*s %s\n", lineNbrIndent, maxLineDigits, "∙∙∙", lineSeparator)
}

// formatUncoveredLine formats a line with the line number in bold and the
// content in red to indicate it is uncovered.
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
			lineRanges, ok := file.NewUncoveredNewLinesGroups[funcName]
			if !ok {
				continue
			}

			displayUncoveredFunctionLines(file, funcName)
			outputUncoveredLinesToFile(outputFile, file.Path, lineRanges)
		}
	}

	return nil
}

// outputUncoveredLineToFile writes the uncovered line range for a given file to
// the specified output file.
func outputUncoveredLineToFile(file *os.File, filePath string, lineRange analyzer.LineRange) {
	if file == nil {
		return
	}

	_, _ = fmt.Fprintf(file, "%s:%d:%d\n", filePath, lineRange.Start, lineRange.End)
}

// outputUncoveredLinesToFile writes multiple uncovered line ranges for a given
// file to the specified output file.
func outputUncoveredLinesToFile(file *os.File, filepath string, lineRanges []analyzer.LineRange) {
	if file == nil {
		return
	}

	for _, lineRange := range lineRanges {
		outputUncoveredLineToFile(file, filepath, lineRange)
	}
}

// padLines takes a list of uncovered lines and pads them with additional lines
// before and after each uncovered line within the function body range. It
// returns a sorted and compacted list of line numbers.
func padLines(lines []int, funcBodyLineStart, funcBodyLineEnd int) []int {
	paddedUncoveredNewLines := make([]int, 0)
	for _, line := range lines {
		for i := line - padLinesBy; i <= line+padLinesBy; i++ {
			if i >= funcBodyLineStart && i <= funcBodyLineEnd {
				paddedUncoveredNewLines = append(paddedUncoveredNewLines, i)
			}
		}
	}

	slices.Sort(paddedUncoveredNewLines)

	return slices.Compact(paddedUncoveredNewLines)
}
