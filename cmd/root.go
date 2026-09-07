package cmd

import (
	"bytes"
	"fmt"
	"os"
	"slices"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/spf13/cobra"

	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"github.com/engmtcdrm/uncloak/internal/config"
)

const (
	floatFormat = "%.2f%%"

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

		for funcName, lineRange := range file.NewUncoveredNewLinesGroups {
			_ = lineRange
			// outputUncoveredLineToStdout(file.Path, lineRange)
			// outputUncoveredLineToStdout2(file.Path, funcName, lineRange)
			outputUncoveredLineToStdout3(file, funcName)
			// outputUncoveredLinetoFile(outputFile, file.Path, lineRange)
		}

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
func outputUncoveredLineToStdout2(filePath string, funcName string, lineRange []analyzer.LineRange) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "  Function: %s\n", pp.Red(funcName))

	for _, lr := range lineRange {
		fmt.Fprintf(&buf, "    %s:%s:%s\n",
			pp.Bold(filePath),
			pp.Redf("%d", lr.Start),
			pp.Redf("%d", lr.End),
		)
		// fmt.Fprintf(&buf, "%s    %s:%s:%s\n",
		// 	pp.RedBgf(" %-3d ", lr.Start),
		// 	pp.Bold(filePath),
		// 	pp.Redf("%d", lr.Start),
		// 	pp.Redf("%d", lr.End),
		// )
		// fmt.Fprintf(&buf, "%s\n", pp.Bg8Bit(239, "─────"))
		// fmt.Fprintf(&buf, "%s", pp.Dimf(" %-3d     covered code already\n", lr.End))
	}

	if len(lineRange) > 0 {
		fmt.Fprintln(&buf)
	}

	fmt.Print(buf.String())
}

func outputUncoveredLineToStdout3(file *analyzer.FileReport, funcName string) {
	f, ok := file.ASTFile.FuncDecls[funcName]
	if !ok {
		return
	}

	codeBodyLineStart := f.StartLine + 1
	if codeBodyLineStart < 1 {
		codeBodyLineStart = 1
	}

	codeBodyLineEnd := f.EndLine - 1
	if codeBodyLineEnd > len(file.ASTFile.Lines) {
		codeBodyLineEnd = len(file.ASTFile.Lines)
	}

	var buf bytes.Buffer

	paddedUncoveredNewLines := make([]int, 0)
	for _, line := range file.UncoveredNewLines {
		for i := line - 3; i <= line+3; i++ {
			if i >= codeBodyLineStart && i <= codeBodyLineEnd {
				paddedUncoveredNewLines = append(paddedUncoveredNewLines, i)
			}
		}
	}

	slices.Sort(paddedUncoveredNewLines)
	paddedUncoveredNewLines = slices.Compact(paddedUncoveredNewLines)
	minLine := slices.Min(paddedUncoveredNewLines)
	maxLine := slices.Max(paddedUncoveredNewLines)

	fmt.Fprintf(&buf, "%s:%s:%s:%s\n\n", pp.Bold(file.Path), pp.Redf("%d", f.StartLine), pp.Redf("%d", f.EndLine), pp.Red(funcName))
	fmt.Fprintf(&buf, "%s %s\n", pp.Dimf("%4d▐", f.StartLine), pp.Dim(file.ASTFile.Lines[f.StartLine-1]))
	// fmt.Fprintf(&buf, "%s %s\n", pp.Bg8Bitf(8, "%4d:", f.StartLine), pp.Bold(file.ASTFile.Lines[f.StartLine-1]))

	if minLine > f.StartLine+1 {
		fmt.Fprintf(&buf, "%s\n", pp.Dim(" ∙∙∙▐"))
		// fmt.Fprintf(&buf, "%s\n", pp.Bg8Bit(0, " ∙∙∙ "))
	}

	for i, lineNbr := range paddedUncoveredNewLines {
		uncovered := slices.Contains(file.UncoveredNewLines, lineNbr)

		if !uncovered {
			fmt.Fprintf(&buf, "%s %s\n", pp.Dimf("%4d▐", lineNbr), pp.Dim(file.ASTFile.Lines[lineNbr-1]))
			// fmt.Fprintf(&buf, "%s %s\n", pp.Bg8Bitf(8, "%4d:", lineNbr), pp.Dim(file.ASTFile.Lines[lineNbr-1]))
		} else {
			fmt.Fprintf(&buf, "%4d%s %s\n", lineNbr, pp.Red("▐"), pp.Red(file.ASTFile.Lines[lineNbr-1]))
			// fmt.Fprintf(&buf, "%s %s\n", pp.Bg8Bitf(1, "%4d:", lineNbr), pp.Red(file.ASTFile.Lines[lineNbr-1]))
		}

		if i < len(paddedUncoveredNewLines)-1 {
			nextLineNbr := paddedUncoveredNewLines[i+1]
			if nextLineNbr > lineNbr+1 {
				fmt.Fprintf(&buf, "%s\n", pp.Dim(" ∙∙∙▐"))
				// fmt.Fprintf(&buf, "%s\n", pp.Bg8Bit(0, " ∙∙∙ "))
			}
		}
	}

	if maxLine+1 < f.EndLine {
		fmt.Fprintf(&buf, "%s\n", pp.Dim(" ∙∙∙▐"))
		// fmt.Fprintf(&buf, "%s\n", pp.Bg8Bit(0, " ∙∙∙ "))
	}

	fmt.Fprintf(&buf, "%s %s\n", pp.Dimf("%4d▐", f.EndLine), pp.Dim(file.ASTFile.Lines[f.EndLine-1]))
	// fmt.Fprintf(&buf, "%s %s\n", pp.Bg8Bitf(8, "%4d:", f.EndLine), pp.Dim(file.ASTFile.Lines[f.EndLine-1]))

	// fmt.Fprintf(&buf, "%s\n", pp.Bg8Bit(239, " ∙∙∙ "))
	// fmt.Fprint(&buf, " ∙∙∙ \n")
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
