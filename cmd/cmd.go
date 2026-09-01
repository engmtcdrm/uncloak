package cmd

import (
	"errors"
	"fmt"

	"github.com/engmtcdrm/uncloak/internal/analyzer"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/header"
	"github.com/spf13/cobra"
)

type cmd struct {
	coverageThreshold float64
	coverageFile      string
	debug             bool
	gitTargetRef      string
	verbose           bool
	output            string
}

// Run executes the command with the provided Cobra command and arguments. It
// handles loading the configuration, validating it, running the code coverage
// analysis, and outputting the results.
func (c *cmd) Run(cmd *cobra.Command, _ []string) error {
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
		errSameBranch := &gitdiff.SameRefError{}

		// If the user provided the same target reference as the current HEAD
		// reference, skip throwing an error because it is not an actual
		// failure. This is primarily here for pipelines to prevent unnecessary
		// failures when the target reference is the same as the current HEAD
		// reference.
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
		return newCoverageThresholdError(report.CoveragePercent(), cfg.CoverageThreshold)
	}

	fmt.Printf("New code coverage %s is equal to or above the minimum required %s\n",
		colors.Greenf(floatFormat, report.CoveragePercent()),
		colors.Greenf(floatFormat, cfg.CoverageThreshold),
	)

	return nil
}

// ValidateFlags checks that the required flags are set and returns an error if
// any validations fail.
func (c *cmd) ValidateFlags(cmd *cobra.Command, _ []string) error {
	if !cmd.Flags().Changed(targetRefFlagName) || c.gitTargetRef == "" {
		return errGitTargetRef
	}

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
