package analyzer

import (
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
)

// Report represents the analysis report for a Go project, including the
// coverage threshold, coverage profile, git diff results, and file-specific
// reports.
type Report struct {
	CoverageThreshold float64
	CoverageProfile   *gocover.Profile
	GitDiffResults    *gitdiff.Results
	Files             []*FileReport
}

// NewReport creates a new [Report] instance with the provided coverage
// threshold, coverage profile, and git diff.
func NewReport(coverageThreshold float64, coverageProfile *gocover.Profile, gitDiff *gitdiff.Results) *Report {
	switch {
	case coverageThreshold < 0.0:
		coverageThreshold = 0.0
	case coverageThreshold > 100.0:
		coverageThreshold = 100.0
	}

	if coverageProfile == nil {
		coverageProfile = &gocover.Profile{}
	}

	if gitDiff == nil {
		gitDiff = &gitdiff.Results{}
	}

	return &Report{
		CoverageThreshold: coverageThreshold,
		CoverageProfile:   coverageProfile,
		GitDiffResults:    gitDiff,
	}
}

// CoveragePercent calculates and returns the coverage percentage of new lines
// in the report.
func (r *Report) CoveragePercent() float64 {
	totalNewLines := r.TotalNewLines()
	if totalNewLines == 0 {
		return 100.0
	}

	return 100.0 * float64(r.TotalCoveredNewLines()) / float64(totalNewLines)
}

// GroupCoveredLines groups the covered new lines in each file into ranges of
// consecutive lines.
func (r *Report) GroupCoveredLines() {
	for _, file := range r.Files {
		file.GroupCoveredLines()
	}
}

// GroupUncoveredLines groups the uncovered new lines in each file into ranges
// of consecutive lines.
func (r *Report) GroupUncoveredLines() {
	for _, file := range r.Files {
		file.GroupUncoveredLines()
	}
}

// HasUncoveredLines checks if there are any uncovered new lines in the report.
func (r *Report) HasUncoveredLines() bool {
	for _, file := range r.Files {
		if len(file.UncoveredNewLines) > 0 {
			return true
		}
	}

	return false
}

// TotalCoveredNewLines calculates and returns the total number of covered new
// lines across all files in the report.
func (r *Report) TotalCoveredNewLines() int {
	var total int
	for _, file := range r.Files {
		total += len(file.CoveredNewLines)
	}

	return total
}

// TotalNewLines calculates and returns the total number of new lines across all
// files in the report.
func (r *Report) TotalNewLines() int {
	var total int
	for _, file := range r.Files {
		total += file.TotalNewLines()
	}

	return total
}
