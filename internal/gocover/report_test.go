package gocover

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [NewReport] function.
func Test_NewReport(t *testing.T) {
	t.Run("return empty report with empty rawTestOutput", func(t *testing.T) {
		report := NewReport("", []byte{})
		require.NotNil(t, report)
		require.Empty(t, report.RawTestOutput)
		require.Empty(t, report.Lines)
		require.Empty(t, report.CoveredLines)
	})

	t.Run("return empty report with nil rawTestOutput", func(t *testing.T) {
		report := NewReport("", nil)
		require.NotNil(t, report)
		require.Empty(t, report.RawTestOutput)
		require.Empty(t, report.Lines)
		require.Empty(t, report.CoveredLines)
	})

	t.Run("return report with rawTestOutput", func(t *testing.T) {
		rawTestOutput := []byte("test output")
		report := NewReport("", rawTestOutput)
		require.NotNil(t, report)
		require.Equal(t, rawTestOutput, report.RawTestOutput)
		require.Empty(t, report.Lines)
		require.Empty(t, report.CoveredLines)
	})
}

// Tests for [Report.Files] function.
func Test_Report_Files(t *testing.T) {
	t.Run("return empty slice when no covered lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		files := report.Files()
		require.Empty(t, files)
	})

	t.Run("return sorted list of file paths", func(t *testing.T) {
		report := NewReport("", []byte{})
		report.CoveredLines["b.go"] = map[int]bool{1: true}
		report.CoveredLines["a.go"] = map[int]bool{1: true}

		files := report.Files()
		require.Equal(t, []string{"a.go", "b.go"}, files)
	})
}

// Tests for [Report.IsInTestCoverage] function.
func Test_Report_IsInTestCoverage(t *testing.T) {
	report := NewReport("", []byte{})

	report.CoveredLines["file.go"] = map[int]bool{
		1: true,
		3: true,
	}

	require.True(t, report.IsInTestCoverage("file.go", 1))
	require.False(t, report.IsInTestCoverage("file.go", 2))
	require.False(t, report.IsInTestCoverage("other.go", 1))
}

// Tests for [Report.expandCoveredLines] function.
func Test_Report_expandCoveredLines(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		report := NewReport("", []byte{})
		line := Line{
			Path:      "file.go",
			StartLine: 1,
			EndLine:   1,
			Count:     1,
		}
		report.expandCoveredLines(line)

		require.True(t, report.CoveredLines["file.go"][1])
	})

	t.Run("multiple lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		line := Line{
			Path:      "file.go",
			StartLine: 1,
			EndLine:   3,
			Count:     1,
		}
		report.expandCoveredLines(line)

		require.True(t, report.CoveredLines["file.go"][1])
		require.True(t, report.CoveredLines["file.go"][2])
		require.True(t, report.CoveredLines["file.go"][3])
	})

	t.Run("uncovered lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		line := Line{
			Path:      "file.go",
			StartLine: 1,
			EndLine:   3,
			Count:     0,
		}
		report.expandCoveredLines(line)

		require.False(t, report.CoveredLines["file.go"][1])
		require.False(t, report.CoveredLines["file.go"][2])
		require.False(t, report.CoveredLines["file.go"][3])
	})

	t.Run("covered and uncovered lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		line1 := Line{
			Path:      "file1.go",
			StartLine: 1,
			EndLine:   2,
			Count:     1,
		}
		line2 := Line{
			Path:      "file1.go",
			StartLine: 3,
			EndLine:   4,
			Count:     0,
		}
		report.expandCoveredLines(line1)
		report.expandCoveredLines(line2)

		require.True(t, report.CoveredLines["file1.go"][1])
		require.True(t, report.CoveredLines["file1.go"][2])
		require.False(t, report.CoveredLines["file1.go"][3])
		require.False(t, report.CoveredLines["file1.go"][4])
	})

	t.Run("covered and uncovered lines in different files", func(t *testing.T) {
		report := NewReport("", []byte{})
		line1 := Line{
			Path:      "file1.go",
			StartLine: 1,
			EndLine:   2,
			Count:     1,
		}
		line2 := Line{
			Path:      "file1.go",
			StartLine: 3,
			EndLine:   4,
			Count:     0,
		}
		line3 := Line{
			Path:      "file2.go",
			StartLine: 1,
			EndLine:   2,
			Count:     1,
		}
		line4 := Line{
			Path:      "file2.go",
			StartLine: 3,
			EndLine:   4,
			Count:     0,
		}
		report.expandCoveredLines(line1)
		report.expandCoveredLines(line2)
		report.expandCoveredLines(line3)
		report.expandCoveredLines(line4)

		require.True(t, report.CoveredLines["file1.go"][1])
		require.True(t, report.CoveredLines["file1.go"][2])
		require.False(t, report.CoveredLines["file1.go"][3])
		require.False(t, report.CoveredLines["file1.go"][4])
		require.True(t, report.CoveredLines["file2.go"][1])
		require.True(t, report.CoveredLines["file2.go"][2])
		require.False(t, report.CoveredLines["file2.go"][3])
		require.False(t, report.CoveredLines["file2.go"][4])
	})
}
