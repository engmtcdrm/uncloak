package gitdiff

var DefaultOptions = Options{
	Unstaged:  true,
	TargetRef: "origin/main",
}

type Options struct {
	Unstaged  bool   `yaml:"unstaged"`   // Whether to include unstaged changes in the git diff
	TargetRef string `yaml:"target-ref"` // Target ref for git diff, e.g., "origin/main" or commit hash.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the git diff command.
func optionsToArgs(opts *Options) []string {
	args := []string{}

	if opts == nil {
		return args
	}

	switch {
	case opts.Unstaged:
		args = append(args, opts.TargetRef, "--", ".")
	default:
		args = append(args, opts.TargetRef+"...")
	}

	args = append(args, "--unified=0")

	return args
}
