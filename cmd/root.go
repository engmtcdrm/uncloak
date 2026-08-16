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
	"github.com/engmtcdrm/uncloak/internal/header"
)

const (
	floatFormat            = "%.2f%%"
	coverageThresholdUsage = "(optional) coverage threshold override. This will also overwrite what is specified in the configuration file"
	debugUsage             = "(optional) enable debug output, e.g. what commands are run"
	gitTargetRefUsage      = "(optional) git target ref to compare against (default: current branch's nearest parent branch)"
	outputUsage            = "(optional) output file to write new code missing coverage to"
	verboseUsage           = "(optional) enable verbose output, e.g. output from go test command"
)

var (
	rootCmd *cobra.Command
)

type cmd struct {
	coverageThreshold float64
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
		RunE:    c.run,
	}

	rootCmd.SilenceUsage = true

	rootCmd.Flags().Float64VarP(&c.coverageThreshold, "coverage-threshold", "c", config.DefaultConfig.CoverageThreshold, coverageThresholdUsage)
	rootCmd.Flags().BoolVarP(&c.debug, "debug", "d", false, debugUsage)
	rootCmd.Flags().StringVarP(&c.gitTargetRef, "target-ref", "t", "", gitTargetRefUsage)
	rootCmd.Flags().StringVarP(&c.output, "output", "o", "", outputUsage)
	rootCmd.Flags().BoolVarP(&c.verbose, "verbose", "v", false, verboseUsage)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func (c *cmd) run(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("new code coverage %s is equal to or above the minimum required %s\n",
		pp.Greenf(floatFormat, report.CoveragePercent()),
		pp.Greenf(floatFormat, cfg.CoverageThreshold),
	)

	return nil
}

// handleFlags updates the configuration based on the command-line flags that
// were set.
func (c *cmd) handleFlags(cfg *config.Config, cmd *cobra.Command) {
	if cmd.Flags().Changed("coverage-threshold") {
		cfg.CoverageThreshold = c.coverageThreshold
	}

	if cmd.Flags().Changed("debug") {
		cfg.Debug = c.debug
	}

	if cmd.Flags().Changed("target-ref") {
		cfg.GitDiffOptions.TargetRef = c.gitTargetRef
	}
}

// outputUncoveredLines writes the uncovered lines from the report to
// [os.Stdout]. If an outputFile is specified, the uncovered lines will also be
// written to that file.
func outputUncoveredLines(report *analyzer.Report, outputFile string) error {
	if !report.HasUncoveredLines() {
		return nil
	}

	var ofile *os.File
	if outputFile != "" {
		var err error
		ofile, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer ofile.Close()
	}

	fmt.Printf("%s\n\n", colors.LightGreen("Missing coverage:"))

	const fileFormat = "%s:%d:%d\n"

	for _, file := range report.Files {
		for _, lineRange := range file.UncoveredNewLineGroups {
			outputUncoveredLineToTerminal(file.Path, lineRange)

			if ofile != nil {
				outputUncoveredLinetoFile(ofile, file.Path, lineRange)
			}
		}

		if len(file.UncoveredNewLineGroups) > 0 {
			fmt.Println()
		}
	}

	return nil
}

func outputUncoveredLineToTerminal(filePath string, lineRange analyzer.LineRange) {
	fmt.Printf("%s:%s:%s\n",
		pp.Bold(filePath),
		pp.Redf("%d", lineRange.Start),
		pp.Redf("%d", lineRange.End),
	)
}

func outputUncoveredLinetoFile(file *os.File, filePath string, lineRange analyzer.LineRange) {
	fmt.Fprintf(file, "%s:%d:%d\n", filePath, lineRange.Start, lineRange.End)
}
