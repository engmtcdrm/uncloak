package gocover

import (
	"fmt"
	"time"
)

// Options represents the configuration options for Go coverage analysis.
type Options struct {
	Rerun   bool          `yaml:"rerun"`   // Whether to rerun tests to generate coverage data.
	Timeout time.Duration `yaml:"timeout"` // Timeout for the coverage analysis in seconds.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the git diff command.
func optionsToArgs(opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	if opts.Timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%v", opts.Timeout))
	}

	return args
}
