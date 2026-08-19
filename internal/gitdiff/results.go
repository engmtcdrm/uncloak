package gitdiff

import "sort"

const (
	OriginMain = "origin/main"
	LocalMain  = "main"
)

type Results struct {
	RootDir  string
	NewLines map[string]map[int]bool
	Command  string
}

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
