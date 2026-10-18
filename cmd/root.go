package cmd

import (
	"github.com/spf13/cobra"

	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/config"
)

const (
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

var rootCmd *cobra.Command

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
