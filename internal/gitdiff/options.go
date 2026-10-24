package gitdiff

const (
	mergeBaseFlag     = "--merge-base" // Flag to use the merge base in git diff command.
	unifiedFlag       = "--unified=0"  // Flag to set the number of context lines in git diff output to 0.
	pathSpecSeparator = "--"           // Separator for pathspec in git diff command.
)

var (
	// DefaultOptions provides the default configuration for the git diff command.
	DefaultOptions = Options{}

	// goFileFilters defines the patterns for Go-related files to include in the
	// git diff command.
	//
	// Use the following pattern to exclude specific files if needed:
	// 	":(exclude)<file pattern>"; Example: ":(exclude)*_test.go"
	goFileFilters = []string{
		"*.go",
		"*go.mod",
		"*go.sum",
	}
)

// Options represents the configuration options for the git diff command.
type Options struct {
	TargetRef string // Target ref for git diff, e.g., "origin/main" or commit hash.
}

// optionsToArgs converts the Options struct into a slice of command-line
// arguments for the git diff command.
func optionsToArgs(opts Options) []string {
	args := []string{
		mergeBaseFlag,
		unifiedFlag,
	}

	if opts.TargetRef == "" {
		opts.TargetRef = headRef
	}

	args = append(args, opts.TargetRef, pathSpecSeparator)
	args = append(args, goFileFilters...)

	return args
}
