package gocover

import "sort"

// Report represents the coverage report for a Go project, including the lines
// of code with coverage information, the lines covered by tests, and the raw
// test output.
type Report struct {
	Lines         []Line                  // The blocks of code with coverage information.
	CoveredLines  map[string]map[int]bool // The lines covered by tests, keyed by file path and line number.
	RawTestOutput []byte                  // The raw output from running tests.
	Command       string                  // The command used to run the tests.
}

// Line represents a block of code in a source file with coverage information.
type Line struct {
	FilePath    string // The absolute file path of the source file.
	Path        string // The relative file path of the source file, relative to the module root.
	Mode        Mode   // The coverage mode for this block.
	StartLine   int    // The starting line number of the code block.
	StartColumn int    // The starting column number of the code block.
	EndLine     int    // The ending line number of the code block.
	EndColumn   int    // The ending column number of the code block.
	Statements  int    // The number of statements in the code block.
	Count       int    // The number of times the code block was executed during testing.
}

// NewReport creates a new [Report] instance with the provide command and raw
// test output.
func NewReport(command string, rawTestOutput []byte) *Report {
	if rawTestOutput == nil {
		rawTestOutput = []byte{}
	}

	return &Report{
		Lines:         []Line{},
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

func (r *Report) expandCoveredLines(cl Line) {
	coverageMap, ok := r.CoveredLines[cl.Path]
	if !ok {
		coverageMap = make(map[int]bool, cl.EndLine-cl.StartLine+1)
		r.CoveredLines[cl.Path] = coverageMap
	}

	for line := cl.StartLine; line <= cl.EndLine; line++ {
		coverageMap[line] = coverageMap[line] || cl.Count > 0
	}
}
