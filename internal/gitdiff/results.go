package gitdiff

import "sort"

const (
	// OriginMain represents the main branch on the remote repository.
	OriginMain = "origin/main"

	// LocalMain represents the main branch in the local repository.
	LocalMain = "main"
)

// Results represents the outcome of a git diff operation, including the root
// directory, new lines added in each file, and the command used to generate
// the diff.
type Results struct {
	RootDir  string
	NewLines map[string]map[int]bool
	Command  string
}

// NewResults creates a new Results instance by determining the git root
// directory and initializing the NewLines map. It returns an error if the git
// root directory cannot be determined.
func NewResults(command string) (*Results, error) {
	rootDir, err := gitRootDir()
	if err != nil {
		return nil, err
	}

	return &Results{
		RootDir:  rootDir,
		NewLines: make(map[string]map[int]bool),
		Command:  command,
	}, nil
}

// Files returns a sorted list of unique file paths that have new lines in the
// git diff.
func (d *Results) Files() []string {
	var files []string
	for file := range d.NewLines {
		files = append(files, file)
	}

	sort.Strings(files)

	return files
}
