package gocover

import (
	"sort"

	"golang.org/x/tools/cover"
)

// Report represents the coverage report for a Go project, including the lines
// of code with coverage information, the lines covered by tests, and the raw
// test output.
type Report struct {
	Blocks        []ProfileBlock          // The blocks of code with coverage information.
	CoveredLines  map[string]map[int]bool // The lines covered by tests, keyed by file path and line number.
	RawTestOutput []byte                  // The raw output from running tests.
	Command       string                  // The command used to run the tests.
}

// ProfileBlock represents a block of code in a source file with coverage
// information.
type ProfileBlock struct {
	cover.ProfileBlock

	FilePath string // The absolute file path of the source file.
	Path     string // The relative file path of the source file, relative to the module root.
	Mode     Mode   // The coverage mode for this block.
}

// NewReport creates a new [Report] instance with the provide command and raw
// test output.
func NewReport(command string, rawTestOutput []byte) *Report {
	if rawTestOutput == nil {
		rawTestOutput = []byte{}
	}

	return &Report{
		Blocks:        []ProfileBlock{},
		CoveredLines:  make(map[string]map[int]bool),
		Command:       command,
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

func (r *Report) expandCoveredLines(cl ProfileBlock) {
	coverageMap, ok := r.CoveredLines[cl.Path]
	if !ok {
		coverageMap = make(map[int]bool, cl.EndLine-cl.StartLine+1)
		r.CoveredLines[cl.Path] = coverageMap
	}

	for line := cl.StartLine; line <= cl.EndLine; line++ {
		coverageMap[line] = coverageMap[line] || cl.Count > 0
	}
}
