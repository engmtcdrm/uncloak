package cmd

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strconv"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/spf13/cobra"

	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"github.com/engmtcdrm/uncloak/internal/config"
)

const (
	floatFormat   = "%.2f%%"
	lineSeparator = "▐"

	coverageFileFlagName = "coverage-file"
	coverageFileUsage    = "(optional) path to the Go coverage file. If not specified, the default is to use the go tool to generate the coverage file"

	coverageThresholdFlagName = "coverage-threshold"
	coverageThresholdUsage    = "(optional) coverage threshold override. This will also overwrite what is specified in the configuration file"

	debugFlagName = "debug"
	debugUsage    = "(optional) enable debug output, e.g. what commands are run"

	outputFlagName = "output"
	outputUsage    = "(optional) file to write new code missing coverage out to"

	targetRefFlagName = "target-ref"
	targetRefUsage    = "(required) git target reference to compare against, e.g. a branch name or commit hash"

	verboseFlagName = "verbose"
	verboseUsage    = "(optional) enable verbose output, e.g. output from go test command. This does not enable verbose go test. Use configuration file to enable verbose go test output"
)

var (
	rootCmd *cobra.Command
)

func init() {
	rootCmd = newRootCmd()
}

func newRootCmd() *cobra.Command {
	c := &cmd{}

	rootCmd = &cobra.Command{
		Use:     app.Name,
		Short:   app.ShortDesc,
		Long:    app.LongDesc,
		Example: app.Name + " -t main",
		Version: app.Version,
		PreRunE: c.ValidateFlags,
		RunE:    c.Run,
	}

	rootCmd.SilenceUsage = true

	rootCmd.Flags().StringVarP(&c.coverageFile, coverageFileFlagName, "C", "", coverageFileUsage)
	rootCmd.Flags().Float64VarP(&c.coverageThreshold, coverageThresholdFlagName, "c", config.DefaultConfig.CoverageThreshold, coverageThresholdUsage)
	rootCmd.Flags().BoolVarP(&c.debug, debugFlagName, "d", false, debugUsage)
	rootCmd.Flags().StringVarP(&c.gitTargetRef, targetRefFlagName, "t", "", targetRefUsage)
	rootCmd.Flags().StringVarP(&c.output, outputFlagName, "o", "", outputUsage)
	rootCmd.Flags().BoolVarP(&c.verbose, verboseFlagName, "v", false, verboseUsage)

	return rootCmd
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
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
			// funcIt := file.ASTFile.FuncDecls[funcName]
			// fmt.Printf("func: %s, start: %d, end: %d\n", funcName, funcIt.StartLine, funcIt.EndLine)
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

	funcBodyLineStart := f.StartLine + 1
	if funcBodyLineStart < 1 {
		funcBodyLineStart = 1
	}

	funcBodyLineEnd := f.EndLine - 1
	if funcBodyLineEnd > len(file.ASTFile.Lines) {
		funcBodyLineEnd = len(file.ASTFile.Lines)
	}

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

func formatDimmedLine(maxLineDigits int, lineNbr int, lineContent string) string {
	return fmt.Sprintf("  %s %s\n", pp.Dimf("%*d %s", maxLineDigits, lineNbr, lineSeparator), pp.Dim(lineContent))
}

func formatElipsisLine(maxLineDigits int) string {
	return fmt.Sprintf("  %*s\n", maxLineDigits, pp.Dimf("∙∙∙ %s", lineSeparator))
}

func formatUncoveredLine(maxLineDigits int, lineNbr int, lineContent string) string {
	return fmt.Sprintf("  %s%s %s\n", pp.Boldf("%*d", maxLineDigits, lineNbr), pp.Redf(" %s", lineSeparator), pp.Red(lineContent))
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
