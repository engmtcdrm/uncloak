package cmd

import (
	"errors"
	"fmt"

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
	coverageThresholdUsage = "(optional) coverage threshold"
	debugUsage             = "(optional) enable debug output, e.g. what commands are run"
	gitTargetRefUsage      = "(optional) git target ref to compare against (default: current branch's nearest parent branch)"
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
}

func init() {
	c := &cmd{}

	rootCmd = &cobra.Command{
		Use:     app.Name,
		Short:   app.ShortDesc,
		Long:    app.LongDesc,
		Example: app.Name,
		Version: getSemVer(app.Version),
		RunE:    c.run,
	}

	rootCmd.SilenceUsage = true

	rootCmd.Flags().Float64VarP(&c.coverageThreshold, "coverage-threshold", "t", config.DefaultConfig.CoverageThreshold, coverageThresholdUsage)
	rootCmd.Flags().BoolVarP(&c.debug, "debug", "d", false, debugUsage)
	rootCmd.Flags().StringVarP(&c.gitTargetRef, "git-target-ref", "T", "", gitTargetRefUsage)
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

	if report.HasUncoveredLines() {
		fmt.Printf("%s\n\n", colors.LightGreen("Missing coverage:"))

		for _, file := range report.Files {
			for _, lineRange := range file.UncoveredNewLineGroups {
				fmt.Printf("%s:%s:%s\n",
					file.Path,
					pp.Redf("%d", lineRange.Start),
					pp.Redf("%d", lineRange.End),
				)
			}
		}

		fmt.Println()
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

func (c *cmd) handleFlags(cfg *config.Config, cmd *cobra.Command) {
	if cmd.Flags().Changed("coverage-threshold") {
		cfg.CoverageThreshold = c.coverageThreshold
	}

	if cmd.Flags().Changed("debug") {
		cfg.Debug = c.debug
	}

	if cmd.Flags().Changed("git-target-ref") {
		cfg.GitDiffOptions.TargetRef = c.gitTargetRef
	}
}
