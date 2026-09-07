package analyzer

import (
	"sort"

	"github.com/engmtcdrm/uncloak/internal/goast"
)

// FileReport represents the coverage report for a single file, including
// covered and uncovered new lines and their respective ranges.
type FileReport struct {
	Path                   string
	CodeLines              CodeLines
	CoveredNewLines        []int       // TODO: Need to add function name association to line
	CoveredNewLineGroups   []LineRange // TODO: Need to add function name association to line ranges
	UncoveredNewLines      []int       // TODO: Need to add function name association to line
	UncoveredNewLineGroups []LineRange // TODO: Need to add function name association to line ranges

	NewCoveredNewLines         map[string][]int
	NewCoveredNewLinesGroups   map[string][]LineRange
	NewUncoveredNewLines       map[string][]int
	NewUncoveredNewLinesGroups map[string][]LineRange

	ASTFile *goast.File
}

type CodeLines []CodeLine

type CodeLine struct {
	Number       int
	Content      string
	FunctionName string
	IsInFunction bool
	IsNew        bool
	IsCovered    bool
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

		NewCoveredNewLines:         make(map[string][]int),
		NewCoveredNewLinesGroups:   make(map[string][]LineRange),
		NewUncoveredNewLines:       make(map[string][]int),
		NewUncoveredNewLinesGroups: make(map[string][]LineRange),
		ASTFile:                    astFile,
	}, nil
}

// GroupCoveredLines groups the covered new lines in the file into ranges of
// consecutive lines.
func (fr *FileReport) GroupCoveredLines() {
	fr.CoveredNewLineGroups = linesToLineRange(fr.CoveredNewLines)

	g := make(map[string][]LineRange)

	for file, lines := range fr.NewCoveredNewLines {
		g[file] = linesToLineRange(lines)
	}
	fr.NewCoveredNewLinesGroups = g
}

// GroupUncoveredLines groups the uncovered new lines in the file into ranges of
// consecutive lines.
func (fr *FileReport) GroupUncoveredLines() {
	fr.UncoveredNewLineGroups = linesToLineRange(fr.UncoveredNewLines)

	g := make(map[string][]LineRange)

	for file, lines := range fr.NewUncoveredNewLines {
		g[file] = linesToLineRange(lines)
	}
	fr.NewUncoveredNewLinesGroups = g
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
