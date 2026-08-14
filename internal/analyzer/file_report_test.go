package analyzer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [NewFileReport] function.
func Test_NewReportFile(t *testing.T) {
	t.Run("should create a new ReportFile instance with the provided path", func(t *testing.T) {
		reportFile := NewFileReport("path/to/file.go")
		require.Equal(t, "path/to/file.go", reportFile.Path)
		require.Equal(t, 0, reportFile.TotalNewLines())
		require.Empty(t, reportFile.CoveredNewLines)
		require.Empty(t, reportFile.UncoveredNewLines)
		require.Empty(t, reportFile.UncoveredNewLineGroups)
	})
}

// Tests for [FileReport.GroupCoveredLines] function.
func Test_FileReport_GroupCoveredLines(t *testing.T) {
	t.Run("should group covered new lines into ranges", func(t *testing.T) {
		reportFile := NewFileReport("path/to/file.go")
		reportFile.CoveredNewLines = []int{10, 11, 12, 14, 15, 17}

		reportFile.GroupCoveredLines()

		expected := []LineRange{
			{Start: 10, End: 12},
			{Start: 14, End: 15},
			{Start: 17, End: 17},
		}
		require.Equal(t, expected, reportFile.CoveredNewLineGroups)
	})
}

// Tests for [FileReport.GroupUncoveredLines] function.
func Test_FileReport_GroupUncoveredLines(t *testing.T) {
	t.Run("should group uncovered new lines into ranges", func(t *testing.T) {
		reportFile := NewFileReport("path/to/file.go")
		reportFile.UncoveredNewLines = []int{1, 2, 3, 5, 6, 8}

		reportFile.GroupUncoveredLines()

		expected := []LineRange{
			{Start: 1, End: 3},
			{Start: 5, End: 6},
			{Start: 8, End: 8},
		}
		require.Equal(t, expected, reportFile.UncoveredNewLineGroups)
	})
}

// Tests for [FileReport.TotalNewLines] function.
func Test_FileReport_TotalNewLines(t *testing.T) {
	t.Run("should return 0 when there are no new lines", func(t *testing.T) {
		reportFile := NewFileReport("path/to/file.go")
		require.Equal(t, 0, reportFile.TotalNewLines())
	})

	t.Run("should return the correct total of new lines", func(t *testing.T) {
		reportFile := NewFileReport("path/to/file.go")
		reportFile.CoveredNewLines = []int{1, 2, 3}
		reportFile.UncoveredNewLines = []int{4, 5}
		require.Equal(t, 5, reportFile.TotalNewLines())
	})
}

// Tests for [linesToLineRange] function.
func Test_linesToLineRange(t *testing.T) {
	t.Run("should return an empty slice if the input is empty", func(t *testing.T) {
		result := linesToLineRange([]int{})
		require.Empty(t, result)
	})

	t.Run("should group consecutive lines into ranges", func(t *testing.T) {
		lines := []int{1, 2, 3, 5, 6, 8}
		expected := []LineRange{
			{Start: 1, End: 3},
			{Start: 5, End: 6},
			{Start: 8, End: 8},
		}
		result := linesToLineRange(lines)
		require.Equal(t, expected, result)
	})

	t.Run("should handle single line input", func(t *testing.T) {
		lines := []int{4}
		expected := []LineRange{
			{Start: 4, End: 4},
		}
		result := linesToLineRange(lines)
		require.Equal(t, expected, result)
	})
}
