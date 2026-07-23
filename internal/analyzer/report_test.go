package analyzer

import (
	"testing"

	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
	"github.com/stretchr/testify/require"
)

// Tests for [NewReport] function.
func Test_NewReport(t *testing.T) {
	t.Run("should create a new Report instance with the provided coverage threshold, coverage profile, and git diff", func(t *testing.T) {
		coverageProfile := &gocover.Profile{}
		gitDiff := &gitdiff.Results{}
		report := NewReport(80.0, coverageProfile, gitDiff)
		require.Equal(t, 80.0, report.CoverageThreshold)
		require.Same(t, coverageProfile, report.CoverageProfile)
		require.Same(t, gitDiff, report.GitDiffResults)
	})

	t.Run("should create a new Report instance with empty coverage profile and git diff", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		require.Equal(t, 80.0, report.CoverageThreshold)
		require.NotNil(t, report.CoverageProfile)
		require.NotNil(t, report.GitDiffResults)
	})

	t.Run("should set threshold to 0.0 if a negative value is provided", func(t *testing.T) {
		report := NewReport(-1.0, nil, nil)
		require.Equal(t, 0.0, report.CoverageThreshold)
	})

	t.Run("should set threshold to 100.0 if a value greater than 100 is provided", func(t *testing.T) {
		report := NewReport(101.0, nil, nil)
		require.Equal(t, 100.0, report.CoverageThreshold)
	})
}

// Tests for [Report.CoveragePercent] method.
func Test_Report_CoveragePercent(t *testing.T) {
	t.Run("should return 100.0 if there are no new lines", func(t *testing.T) {
		report := &Report{}
		require.Equal(t, 100.0, report.CoveragePercent())
	})

	t.Run("should return the correct coverage percentage", func(t *testing.T) {
		report := &Report{
			Files: []*FileReport{
				{CoveredNewLines: []int{1, 2, 3, 4, 5, 6, 7, 8}, UncoveredNewLines: []int{9, 10}},
				{CoveredNewLines: []int{1, 2, 3}, UncoveredNewLines: []int{4, 5}},
			},
		}
		expectedCoverage := 100 * (float64(8+3) / float64(10+5))
		require.Equal(t, expectedCoverage, report.CoveragePercent())
	})
}

// Tests for [Report.GroupCoveredLines] method.
func Test_Report_GroupCoveredLines(t *testing.T) {
	t.Run("empty report should produce empty CoveredNewLineGroups", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)

		report.GroupCoveredLines()

		for _, file := range report.Files {
			require.Empty(t, file.CoveredNewLineGroups)
		}
	})

	t.Run("populated file CoveredNewLines should produce CoveredNewLineGroups", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)

		reportFile := NewFileReport("test.go")
		reportFile.CoveredNewLines = []int{1, 2, 3, 5, 6, 8}

		report.Files = append(report.Files, reportFile)

		report.GroupCoveredLines()

		for _, file := range report.Files {
			require.NotEmpty(t, file.CoveredNewLineGroups)
			require.Len(t, file.CoveredNewLineGroups, 3)
		}
	})
}

// Tests for [Report.GroupUncoveredLines] method.
func Test_Report_GroupUncoveredLines(t *testing.T) {
	t.Run("empty report should produce empty UncoveredNewLineGroups", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)

		report.GroupUncoveredLines()

		for _, file := range report.Files {
			require.Empty(t, file.UncoveredNewLineGroups)
		}
	})

	t.Run("populated file UncoveredNewLines should produce UncoveredNewLineGroups", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)

		reportFile := NewFileReport("test.go")
		reportFile.UncoveredNewLines = []int{1, 2, 3, 5, 6, 8}

		report.Files = append(report.Files, reportFile)

		report.GroupUncoveredLines()

		for _, file := range report.Files {
			require.NotEmpty(t, file.UncoveredNewLineGroups)
			require.Len(t, file.UncoveredNewLineGroups, 3)
		}
	})
}

// Tests for [Report.HasUncoveredLines] method.
func Test_Report_HasUncoveredLines(t *testing.T) {
	t.Run("should return false if there are no files in the report", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		require.False(t, report.HasUncoveredLines())
	})

	t.Run("should return false if all files have no uncovered new lines", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		report.Files = append(report.Files, &FileReport{UncoveredNewLines: []int{}})
		report.Files = append(report.Files, &FileReport{UncoveredNewLines: []int{}})
		require.False(t, report.HasUncoveredLines())
	})

	t.Run("should return true if any file has uncovered new lines", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		report.Files = append(report.Files, &FileReport{UncoveredNewLines: []int{}})
		report.Files = append(report.Files, &FileReport{UncoveredNewLines: []int{1}})
		require.True(t, report.HasUncoveredLines())
	})
}

// Tests for [Report.TotalCoveredNewLines] method.
func Test_Report_TotalCoveredNewLines(t *testing.T) {
	t.Run("should return 0 if there are no files in the report", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		require.Equal(t, 0, report.TotalCoveredNewLines())
	})

	t.Run("should return the total number of covered new lines across all files", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		report.Files = append(report.Files, &FileReport{CoveredNewLines: []int{1, 2, 3, 4, 5}})
		report.Files = append(report.Files, &FileReport{CoveredNewLines: []int{1, 2, 3}})
		require.Equal(t, 8, report.TotalCoveredNewLines())
	})
}

// Tests for [Report.TotalNewLines] method.
func Test_Report_TotalNewLines(t *testing.T) {
	t.Run("should return 0 if there are no files in the report", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		require.Equal(t, 0, report.TotalNewLines())
	})

	t.Run("should return the total number of new lines across all files", func(t *testing.T) {
		report := NewReport(80.0, nil, nil)
		report.Files = append(report.Files, &FileReport{CoveredNewLines: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}})
		report.Files = append(report.Files, &FileReport{CoveredNewLines: []int{1, 2, 3, 4, 5}})
		require.Equal(t, 15, report.TotalNewLines())
	})
}
