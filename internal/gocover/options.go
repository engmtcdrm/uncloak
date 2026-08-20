package gocover

import (
	"fmt"
	"time"
)

// Options represents the configuration options for Go coverage analysis.
type Options struct {
	Count   int           `yaml:"count"`   // Number of times to run the tests for coverage analysis.
	Timeout time.Duration `yaml:"timeout"` // Timeout for the coverage analysis in seconds.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the git diff command.
func optionsToArgs(opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	if opts.Count > 0 {
		args = append(args, "--count", fmt.Sprintf("%d", opts.Count))
	}

	if opts.Timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%v", opts.Timeout))
	}

	return args
}
