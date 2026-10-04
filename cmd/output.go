package cmd

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"github.com/engmtcdrm/uncloak/internal/goast"
)

const (
	ellipsis        = "∙∙∙" // Ellipsis used to indicate skipped lines
	lineNbrIndentBy = 2     // Number of spaces to indent the line number column
	lineSeparator   = "▐"   // Separator between the line number and the content
	padLinesBy      = 3     // Number of lines to pad before and after each uncovered line
)

var (
	ellipsisLen   = utf8.RuneCountInString(ellipsis)     // Length of the ellipsis string
	lineNbrIndent = strings.Repeat(" ", lineNbrIndentBy) // String of spaces used to indent the line number column
)

// displayUncoveredFunctionLines displays the uncovered lines for a given
// function within a file to [os.Stdout].
func displayUncoveredFunctionLines(fileReport *analyzer.FileReport, funcName string) {
	funcDecl, ok := fileReport.ASTFile.FuncDecls[funcName]
	if !ok {
		return
	}

	funcUncoveredNewLines, ok := fileReport.FuncUncoveredNewLines[funcName]
	if !ok {
		return
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s\n\n", pp.Boldf("%s:%d:%d", fileReport.Path, funcDecl.Start, funcDecl.End))

	bodyStart := max(funcDecl.Body.Start, 1)
	bodyEnd := min(funcDecl.Body.End, len(fileReport.ASTFile.Lines))
	maxLineDigits := max(len(strconv.Itoa(funcDecl.End)), ellipsisLen)

	// If the function body is a single line, display it and return immediately.
	if bodyStart == bodyEnd {
		paddedUncoveredNewLines := padLines(funcUncoveredNewLines, bodyStart, bodyEnd)

		fmt.Fprintln(&buf, formatLines(fileReport, bodyStart, maxLineDigits, paddedUncoveredNewLines, 0))

		fmt.Print(buf.String())

		return
	}

	// Remove first and last lines from the uncovered lines as they are part
	// of the function signature and end.
	paddedUncoveredNewLines := padLines(funcUncoveredNewLines, bodyStart+1, bodyEnd-1)

	minLine := 0
	maxLine := 0

	if len(paddedUncoveredNewLines) > 0 {
		minLine = slices.Min(paddedUncoveredNewLines)
		maxLine = slices.Max(paddedUncoveredNewLines)
	}

	// Always display the function signature line(s).
	fmt.Fprint(&buf, formatFuncSignature(maxLineDigits, fileReport, funcDecl))

	// If the first uncovered line is not immediately after the function
	// signature, add an ellipsis line.
	if minLine > funcDecl.Body.Start+1 {
		fmt.Fprint(&buf, formatEllipsisLine(maxLineDigits))
	}

	for i, lineNbr := range paddedUncoveredNewLines {
		fmt.Fprint(&buf, formatLines(fileReport, lineNbr, maxLineDigits, paddedUncoveredNewLines, i))
	}

	// If the last uncovered line is not immediately before the function end,
	// add an ellipsis line.
	if maxLine+1 < funcDecl.End {
		fmt.Fprint(&buf, formatEllipsisLine(maxLineDigits))
	}

	// Always display the last line of the function as dimmed.
	fmt.Fprint(&buf, formatDimmedLine(maxLineDigits, funcDecl.End, fileReport.ASTFile.Lines[funcDecl.End-1]))

	buf2 := buf.String()
	_ = buf2
	fmt.Fprintln(&buf)

	fmt.Print(buf.String())
}

// displayUncoveredLine displays the uncovered line range for a given file to
// [os.Stdout].
func displayUncoveredLine(filePath string, lineRange analyzer.LineRange) {
	fmt.Printf("%s:%s:%s\n",
		pp.Bold(filePath),
		pp.Redf("%d", lineRange.Start),
		pp.Redf("%d", lineRange.End),
	)
}

// displayUncoveredLines displays the uncovered line ranges for a given file to
// [os.Stdout].
func displayUncoveredLines(filePath string, lineRanges []analyzer.LineRange) {
	for _, lineRange := range lineRanges {
		displayUncoveredLine(filePath, lineRange)
	}
}

// formatDimmedLine formats a line with dimmed text for the line number and
// content.
func formatDimmedLine(maxLineDigits int, lineNbr int, lineContent string) string {
	return pp.Dimf("%s%*d %s %s\n",
		lineNbrIndent,
		maxLineDigits,
		lineNbr,
		lineSeparator,
		lineContent,
	)
}

// formatEllipsisLine formats an ellipsis line with dimmed text for the line
// number column.
func formatEllipsisLine(maxLineDigits int) string {
	return pp.Dimf("%s%*s %s\n",
		lineNbrIndent,
		maxLineDigits,
		ellipsis,
		lineSeparator,
	)
}

// formatFuncSignature formats the function signature lines with dimmed text.
func formatFuncSignature(maxLineDigits int, fileReport *analyzer.FileReport, funcDecl *goast.FuncDecl) string {
	var buf bytes.Buffer

	signatureLen := funcDecl.Body.Start - funcDecl.Start + 1

	for line := range signatureLen {
		fmt.Fprint(&buf, formatDimmedLine(
			maxLineDigits,
			funcDecl.Start+line,
			fileReport.ASTFile.Lines[funcDecl.Start+line-1]),
		)
	}

	return buf.String()
}

// formatLines formats the lines of a file, highlighting uncovered lines and
// adding ellipsis for skipped lines.
func formatLines(fileReport *analyzer.FileReport, lineNbr int, maxLineDigits int, paddedUncoveredNewLines []int, idx int) string {
	var buf bytes.Buffer
	uncovered := slices.Contains(fileReport.UncoveredNewLines, lineNbr)

	if !uncovered {
		fmt.Fprint(&buf, formatDimmedLine(maxLineDigits, lineNbr, fileReport.ASTFile.Lines[lineNbr-1]))
	} else {
		fmt.Fprint(&buf, formatUncoveredLine(maxLineDigits, lineNbr, fileReport.ASTFile.Lines[lineNbr-1]))
	}

	// If there is a gap between the current line and the next line, add an
	// ellipsis line.
	if idx < len(paddedUncoveredNewLines)-1 {
		nextLineNbr := paddedUncoveredNewLines[idx+1]
		if nextLineNbr > lineNbr+1 {
			fmt.Fprint(&buf, formatEllipsisLine(maxLineDigits))
		}
	}

	return buf.String()
}

// formatUncoveredLine formats a line with the line number in bold and the
// content in red to indicate it is uncovered.
func formatUncoveredLine(maxLineDigits int, lineNbr int, lineContent string) string {
	return fmt.Sprintf("%s%s%s\n",
		lineNbrIndent,
		pp.Boldf("%*d", maxLineDigits, lineNbr),
		pp.Redf(" %s %s", lineSeparator, lineContent),
	)
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
		if len(file.FuncUncoveredNewLinesGroups) == 0 {
			continue
		}

		for _, funcName := range file.ASTFile.FuncOrder {
			lineRanges, ok := file.FuncUncoveredNewLinesGroups[funcName]
			if !ok {
				continue
			}

			displayUncoveredFunctionLines(file, funcName)
			// displayUncoveredLines(file.Path, lineRanges)
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
	if len(lines) == 0 {
		return nil
	}

	var paddedUncoveredNewLines []int
	for _, line := range lines {
		for i := line - padLinesBy; i <= line+padLinesBy; i++ {
			if i < funcBodyLineStart || i > funcBodyLineEnd {
				continue
			}

			paddedUncoveredNewLines = append(paddedUncoveredNewLines, i)
		}
	}

	slices.Sort(paddedUncoveredNewLines)

	return slices.Compact(paddedUncoveredNewLines)
}
