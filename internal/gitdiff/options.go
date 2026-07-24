package gitdiff

var DefaultOptions = Options{}

type Options struct {
	TargetRef string // Target ref for git diff, e.g., "origin/main" or commit hash.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the git diff command.
func optionsToArgs(opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	if opts.TargetRef == "" {
		opts.TargetRef = findNearestParent()
	}

	args = append(args, opts.TargetRef, "--", ".", "--unified=0")

	return args
}
