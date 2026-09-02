package cmd

import (
	"fmt"
	"os"

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
	targetRefUsage    = "(required) git target reference to compare against, e.g. a branch name or commit SHA"

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
		for _, lineRange := range file.UncoveredNewLineGroups {
			outputUncoveredLineToStdout(file.Path, lineRange)
			outputUncoveredLinetoFile(outputFile, file.Path, lineRange)
		}

		if len(file.UncoveredNewLineGroups) > 0 {
			fmt.Println()
		}
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
