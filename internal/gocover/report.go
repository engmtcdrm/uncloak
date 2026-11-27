package gocover

import (
	"sort"

	"golang.org/x/tools/cover"
)

// Report represents the coverage report for a Go project, including the lines
// of code with coverage information, the lines covered by tests, and the raw
// test output.
type Report struct {
	Blocks        []cover.ProfileBlock    // The blocks of code with coverage information.
	Command       string                  // The command used to run the tests.
	CoveredLines  map[string]map[int]bool // The lines covered by tests, keyed by file path and line number.
	Profiles      []*Profile              // The coverage profiles for the source files.
	RawTestOutput []byte                  // The raw output from running tests.

}

// Profile represents the profiling data for a specific file.
type Profile struct {
	cover.Profile

	NoModuleFileName string // The file path of the source file relative to the module root, without the module prefix.
}

// NewReport creates a new [Report] instance with the provide command and raw
// test output.
func NewReport(command string, rawTestOutput []byte) *Report {
	if rawTestOutput == nil {
		rawTestOutput = []byte{}
	}

	return &Report{
		Blocks:        []cover.ProfileBlock{},
		Command:       command,
		CoveredLines:  make(map[string]map[int]bool),
		Profiles:      []*Profile{},
		RawTestOutput: rawTestOutput,
	}
}

// Files returns a sorted list of unique file paths that have coverage data in
// the report.
func (r *Report) Files() []string {
	files := make([]string, 0, len(r.CoveredLines))
	for file := range r.CoveredLines {
		files = append(files, file)
	}

	sort.Strings(files)

	return files
}

// IsInTestCoverage checks if a specific line in a file is covered by tests
// according to the report's coverage data.
func (r *Report) IsInTestCoverage(file string, line int) bool {
	_, ok := r.CoveredLines[file][line]
	return ok
}

func (r *Report) expandCoveredLines(fileName string, pb cover.ProfileBlock) {
	coverageMap, ok := r.CoveredLines[fileName]
	if !ok {
		coverageMap = make(map[int]bool, pb.EndLine-pb.StartLine+1)
		r.CoveredLines[fileName] = coverageMap
	}

	for line := pb.StartLine; line <= pb.EndLine; line++ {
		coverageMap[line] = coverageMap[line] || pb.Count > 0
	}
}
