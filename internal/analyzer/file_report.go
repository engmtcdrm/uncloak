package analyzer

import (
	"sort"

	"github.com/engmtcdrm/uncloak/internal/goast"
)

// FileReport represents the coverage report for a single file, including
// covered and uncovered new lines and their respective ranges.
type FileReport struct {
	Path                   string      // Path to the file.
	CodeLines              CodeLines   // All code lines in the file.
	CoveredNewLines        []int       // Covered new lines in the file.
	CoveredNewLineGroups   []LineRange // Groups of consecutive covered new lines.
	UncoveredNewLines      []int       // Uncovered new lines in the file.
	UncoveredNewLineGroups []LineRange // Groups of consecutive uncovered new lines.

	FuncCoveredNewLines         map[string][]int       // Covered new lines per function.
	FuncCoveredNewLinesGroups   map[string][]LineRange // Groups of consecutive covered new lines per function.
	FuncUncoveredNewLines       map[string][]int       // Uncovered new lines per function.
	FuncUncoveredNewLinesGroups map[string][]LineRange // Groups of consecutive uncovered new lines per function.

	ASTFile *goast.File
}

// CodeLines represents a collection of [CodeLine].
type CodeLines []CodeLine

// CodeLine represents a single line of code in a file, including its line
// number, content, associated function, and coverage status.
type CodeLine struct {
	Number       int    // Line number of the code line
	Content      string // Content of the code line
	FunctionName string // Name of the function the code line belongs to
	IsInFunction bool   // Indicates if the code line is inside a function
	IsNew        bool   // Indicates if the code line is new
	IsCovered    bool   // Indicates if the code line is covered
}

// LineRange represents a continuous range of lines in a file, with a start and
// end line number.
type LineRange struct {
	Start int
	End   int
}

// NewFileReport creates a new [FileReport] instance for the given file path.
func NewFileReport(path string) (*FileReport, error) {
	astFile, err := goast.Parse(path)
	if err != nil {
		return nil, err
	}

	return &FileReport{
		Path:                   path,
		CoveredNewLines:        make([]int, 0),
		CoveredNewLineGroups:   make([]LineRange, 0),
		UncoveredNewLines:      make([]int, 0),
		UncoveredNewLineGroups: make([]LineRange, 0),

		FuncCoveredNewLines:         make(map[string][]int),
		FuncCoveredNewLinesGroups:   make(map[string][]LineRange),
		FuncUncoveredNewLines:       make(map[string][]int),
		FuncUncoveredNewLinesGroups: make(map[string][]LineRange),
		ASTFile:                     astFile,
	}, nil
}

// GroupCoveredLines groups the covered new lines in the file into ranges of
// consecutive lines.
func (fr *FileReport) GroupCoveredLines() {
	fr.CoveredNewLineGroups = linesToLineRange(fr.CoveredNewLines)

	lineRanges := make(map[string][]LineRange, len(fr.FuncCoveredNewLines))

	for file, lines := range fr.FuncCoveredNewLines {
		lineRanges[file] = linesToLineRange(lines)
	}

	fr.FuncCoveredNewLinesGroups = lineRanges
}

// GroupUncoveredLines groups the uncovered new lines in the file into ranges of
// consecutive lines.
func (fr *FileReport) GroupUncoveredLines() {
	fr.UncoveredNewLineGroups = linesToLineRange(fr.UncoveredNewLines)

	lineRanges := make(map[string][]LineRange, len(fr.FuncUncoveredNewLines))

	for file, lines := range fr.FuncUncoveredNewLines {
		lineRanges[file] = linesToLineRange(lines)
	}

	fr.FuncUncoveredNewLinesGroups = lineRanges
}

// TotalNewLines returns the total number of new lines in the file report, which
// is the sum of covered and uncovered new lines.
func (fr *FileReport) TotalNewLines() int {
	return len(fr.CoveredNewLines) + len(fr.UncoveredNewLines)
}

// linesToLineRange converts lines into continuous ranges. For example,
// [1, 2, 3, 5, 6] becomes [[1, 3], [5, 6]].
func linesToLineRange(lines []int) []LineRange {
	if len(lines) == 0 {
		return nil
	}

	sort.Ints(lines)

	var ranges []LineRange
	start := lines[0]
	end := lines[0]

	for _, line := range lines[1:] {
		if line == end+1 {
			end = line
			continue
		}

		ranges = append(ranges, LineRange{Start: start, End: end})
		start = line
		end = line
	}

	ranges = append(ranges, LineRange{Start: start, End: end})

	return ranges
}
