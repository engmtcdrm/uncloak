package gitdiff

import "context"

// DefaultOptions provides the default configuration for the git diff command.
var DefaultOptions = Options{}

// Options represents the configuration options for the git diff command.
type Options struct {
	TargetRef string // Target ref for git diff, e.g., "origin/main" or commit hash.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the git diff command.
func optionsToArgs(ctx context.Context, opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	if opts.TargetRef == "" {
		opts.TargetRef = findNearestParent(ctx)
	}

	args = append(args, opts.TargetRef, "--", ".", "--unified=0")

	return args
}
