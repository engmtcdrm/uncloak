package cmd

import (
	"errors"
	"fmt"
	"os"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/spf13/cobra"

	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/header"
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
	targetRefUsage    = "(required) git target ref to compare against, e.g. 'main'"

	verboseFlagName = "verbose"
	verboseUsage    = "(optional) enable verbose output, e.g. output from go test command. This does not enable verbose go test. Use configuration file to enable verbose go test output"
)

var (
	rootCmd *cobra.Command

	errGitTargetRef = errors.New("flag -t/--target-ref is required. This should be the target ref to compare against, e.g. 'main'")
)

type cmd struct {
	coverageThreshold float64
	coverageFile      string
	debug             bool
	gitTargetRef      string
	verbose           bool
	output            string
}

func init() {
	c := &cmd{}

	rootCmd = &cobra.Command{
		Use:     app.Name,
		Short:   app.ShortDesc,
		Long:    app.LongDesc,
		Example: app.Name,
		Version: app.Version,
		PreRunE: c.validateFlags,
		RunE:    c.run,
	}

	rootCmd.SilenceUsage = true

	rootCmd.Flags().StringVarP(&c.coverageFile, coverageFileFlagName, "C", "", coverageFileUsage)
	rootCmd.Flags().Float64VarP(&c.coverageThreshold, coverageThresholdFlagName, "c", config.DefaultConfig.CoverageThreshold, coverageThresholdUsage)
	rootCmd.Flags().BoolVarP(&c.debug, debugFlagName, "d", false, debugUsage)
	rootCmd.Flags().StringVarP(&c.gitTargetRef, targetRefFlagName, "t", "", targetRefUsage)
	rootCmd.Flags().StringVarP(&c.output, outputFlagName, "o", "", outputUsage)
	rootCmd.Flags().BoolVarP(&c.verbose, verboseFlagName, "v", false, verboseUsage)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func (c *cmd) run(cmd *cobra.Command, _ []string) error {
	header.PrintHeader()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	c.handleFlags(cfg, cmd)

	if err := config.Validate(cfg); err != nil {
		return err
	}

	report, err := analyzer.NewCodeCoverage(cfg)
	if err != nil && !errors.Is(err, analyzer.ErrBelowThreshold) {
		errSameBranch := &gitdiff.SameBranchError{}

		// If the user provided the same target ref branch as the current
		// branch, skip throwing an error because it is not an actual failure.
		// This is primarily here for pipelines to prevent unnecessary failures
		// when the target ref is the same as the current branch.
		if errors.As(err, &errSameBranch) {
			return nil
		}

		return err
	}

	if c.verbose {
		fmt.Printf("%s\n\n", colors.LightGreen("Go Test Output:"))
		fmt.Println(string(report.CoverageProfile.RawTestOutput))
	}

	if err := outputUncoveredLines(report, c.output); err != nil {
		return err
	}

	if err != nil {
		err = fmt.Errorf("new code coverage %s is below the minimum required %s",
			pp.Redf(floatFormat, report.CoveragePercent()),
			pp.Redf(floatFormat, cfg.CoverageThreshold),
		)
		return err
	}

	fmt.Printf("New code coverage %s is equal to or above the minimum required %s\n",
		colors.Greenf(floatFormat, report.CoveragePercent()),
		colors.Greenf(floatFormat, cfg.CoverageThreshold),
	)

	return nil
}

// handleFlags updates the configuration based on the command-line flags that
// were set.
func (c *cmd) handleFlags(cfg *config.Config, cmd *cobra.Command) {
	if cmd.Flags().Changed(coverageFileFlagName) {
		cfg.CoverageFile = c.coverageFile
	}

	if cmd.Flags().Changed(coverageThresholdFlagName) {
		cfg.CoverageThreshold = c.coverageThreshold
	}

	if cmd.Flags().Changed(debugFlagName) {
		cfg.Debug = c.debug
	}

	if cmd.Flags().Changed(targetRefFlagName) {
		cfg.GitDiffOptions.TargetRef = c.gitTargetRef
	}
}

// validateFlags checks that the required flags are set and returns an error if
// any validations fail.
func (c *cmd) validateFlags(cmd *cobra.Command, _ []string) error {
	if !cmd.Flags().Changed(targetRefFlagName) {
		return errGitTargetRef
	}

	return nil
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
