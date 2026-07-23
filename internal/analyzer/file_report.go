package analyzer

import "sort"

type FileReport struct {
	Path                   string
	CoveredNewLines        []int
	CoveredNewLineGroups   []LineRange
	UncoveredNewLines      []int
	UncoveredNewLineGroups []LineRange
}

type LineRange struct {
	Start int
	End   int
}

// NewFileReport creates a new [FileReport] instance for the given file path.
func NewFileReport(path string) *FileReport {
	return &FileReport{
		Path:                   path,
		CoveredNewLines:        make([]int, 0),
		CoveredNewLineGroups:   make([]LineRange, 0),
		UncoveredNewLines:      make([]int, 0),
		UncoveredNewLineGroups: make([]LineRange, 0),
	}
}

// GroupCoveredLines groups the covered new lines in the file into ranges of
// consecutive lines.
func (fr *FileReport) GroupCoveredLines() {
	fr.CoveredNewLineGroups = linesToLineRange(fr.CoveredNewLines)
}

// GroupUncoveredLines groups the uncovered new lines in the file into ranges of
// consecutive lines.
func (fr *FileReport) GroupUncoveredLines() {
	fr.UncoveredNewLineGroups = linesToLineRange(fr.UncoveredNewLines)
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
