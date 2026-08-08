package gocover

import "sort"

type Profile struct {
	Mode          Mode
	Lines         []Line
	CoveredLines  map[string]map[int]bool
	RawTestOutput []byte
	Command       string
}

type Line struct {
	FilePath    string // The absolute file path of the source file.
	Path        string // The relative file path of the source file, relative to the module root.
	StartLine   int    // The starting line number of the code block.
	StartColumn int    // The starting column number of the code block.
	EndLine     int    // The ending line number of the code block.
	EndColumn   int    // The ending column number of the code block.
	Statements  int    // The number of statements in the code block.
	Count       int    // The number of times the code block was executed during testing.
}

// NewProfile creates a new Profile instance with the provided raw test output.
func NewProfile(command string, rawTestOutput []byte) *Profile {
	if rawTestOutput == nil {
		rawTestOutput = []byte{}
	}

	return &Profile{
		Lines:         []Line{},
		CoveredLines:  make(map[string]map[int]bool),
		Command:       command,
		RawTestOutput: rawTestOutput,
	}
}

// Files returns a sorted list of unique file paths that have coverage data in the profile.
func (cp *Profile) Files() []string {
	files := make([]string, 0, len(cp.CoveredLines))
	for file := range cp.CoveredLines {
		files = append(files, file)
	}

	sort.Strings(files)

	return files
}

// IsInTestCoverage checks if a specific line in a file is covered by tests
// according to the profile's coverage data.
func (cp *Profile) IsInTestCoverage(file string, line int) bool {
	_, ok := cp.CoveredLines[file][line]
	return ok
}

func (cp *Profile) expandCoveredLines(cl Line) {
	coverageMap, ok := cp.CoveredLines[cl.Path]
	if !ok {
		coverageMap = make(map[int]bool, cl.EndLine-cl.StartLine+1)
		cp.CoveredLines[cl.Path] = coverageMap
	}

	for line := cl.StartLine; line <= cl.EndLine; line++ {
		coverageMap[line] = cl.Count > 0
	}
}
