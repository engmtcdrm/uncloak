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
		for lineNbr := funcDecl.Start; lineNbr <= bodyEnd; lineNbr++ {
			fmt.Fprint(&buf, formatLines(fileReport, lineNbr, maxLineDigits, nil, 0))
		}

		fmt.Fprintln(&buf)

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

// outputUncoveredLine writes the uncovered lines for a single file to
// [os.Stdout]. If outputFile is not nil, the uncovered lines will also be
// written to that file. If linesOnly is true the output will only include
// uncovered lines without detailed context.
func outputUncoveredLine(file *analyzer.FileReport, outputFile *os.File, linesOnly bool) {
	if file == nil {
		return
	}

	// Report unmapped uncovered lines even when an AST is available.
	unmappedLines := unmappedUncoveredLineRanges(file)
	displayUncoveredLines(file.Path, unmappedLines)
	outputUncoveredLinesToFile(outputFile, file.Path, unmappedLines)

	if file.ASTFile == nil {
		return
	}

	for _, funcName := range file.ASTFile.FuncOrder {
		lineRanges, ok := file.FuncUncoveredNewLinesGroups[funcName]
		if !ok {
			continue
		}

		switch linesOnly {
		case true:
			displayUncoveredLines(file.Path, lineRanges)
		default:
			displayUncoveredFunctionLines(file, funcName)
		}

		outputUncoveredLinesToFile(outputFile, file.Path, lineRanges)
	}

	if linesOnly {
		fmt.Println()
	}
}

// outputUncoveredLines writes the uncovered lines from the report to
// [os.Stdout]. If an output file path is specified, the uncovered lines will
// also be written to that file. If linesOnly is true the output will only
// include uncovered lines without detailed context.
func outputUncoveredLines(report *analyzer.Report, outputFilePath string, linesOnly bool, filesToFilter ...string) error {
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

	if len(filesToFilter) > 0 && filesToFilter[0] != "" {
		fmt.Printf("Filtering missing coverage by files: %v\n\n", filesToFilter)
	}

	uncoveredFiles := report.UncoveredFiles()

	for _, file := range uncoveredFiles {
		if len(filesToFilter) > 0 && filesToFilter[0] != "" && !isFilterFile(file.Path, filesToFilter) {
			continue
		}

		outputUncoveredLine(file, outputFile, linesOnly)
	}

	return nil
}

// isFilterFile checks if the given file matches any of the exclusion
// patterns in the config.
func isFilterFile(file string, files []string) bool {
	for _, f := range files {
		if file == f {
			return true
		}
	}

	return false
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

// unmappedUncoveredLineRanges returns the line ranges of uncovered lines that
// are not mapped to any function in the given file report.
func unmappedUncoveredLineRanges(file *analyzer.FileReport) []analyzer.LineRange {
	mappedLines := make(map[int]struct{})
	for _, lines := range file.FuncUncoveredNewLines {
		for _, line := range lines {
			mappedLines[line] = struct{}{}
		}
	}

	var unmappedLines []int
	for _, line := range file.UncoveredNewLines {
		if _, ok := mappedLines[line]; !ok {
			unmappedLines = append(unmappedLines, line)
		}
	}

	slices.Sort(unmappedLines)
	if len(unmappedLines) == 0 {
		return nil
	}

	lineRanges := make([]analyzer.LineRange, 0, len(unmappedLines))
	start := unmappedLines[0]
	end := start
	for _, line := range unmappedLines[1:] {
		if line == end+1 {
			end = line
			continue
		}

		lineRanges = append(lineRanges, analyzer.LineRange{Start: start, End: end})
		start = line
		end = line
	}

	return append(lineRanges, analyzer.LineRange{Start: start, End: end})
}
