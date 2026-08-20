package gocover

import (
	"fmt"
	"time"
)

// DefaultOptions provides the default configuration for the go test command.
var DefaultOptions = Options{}

// Options represents the configuration options for Go coverage analysis.
type Options struct {
	Count   int           `yaml:"count"`   // Number of times to run the tests for coverage analysis.
	Timeout time.Duration `yaml:"timeout"` // Timeout duration for the coverage analysis.
	Verbose bool          `yaml:"verbose"` // Enable verbose output for the coverage analysis.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the go test command.
func optionsToArgs(opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	if opts.Count > 0 {
		args = append(args, "-count", fmt.Sprintf("%d", opts.Count))
	}

	if opts.Timeout > 0 {
		args = append(args, "-timeout", fmt.Sprintf("%v", opts.Timeout))
	}

	if opts.Verbose {
		args = append(args, "-v")
	}

	return args
}
