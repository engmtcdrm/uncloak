package gocover

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/cover"
)

// Tests for [NewReport] function.
func Test_NewReport(t *testing.T) {
	t.Run("return empty report with empty rawTestOutput", func(t *testing.T) {
		report := NewReport("", []byte{})
		require.NotNil(t, report)
		require.Empty(t, report.RawTestOutput)
		require.Empty(t, report.Profiles)
		require.Empty(t, report.CoveredLines)
	})

	t.Run("return empty report with nil rawTestOutput", func(t *testing.T) {
		report := NewReport("", nil)
		require.NotNil(t, report)
		require.Empty(t, report.RawTestOutput)
		require.Empty(t, report.Profiles)
		require.Empty(t, report.CoveredLines)
	})

	t.Run("return report with rawTestOutput", func(t *testing.T) {
		rawTestOutput := []byte("test output")
		report := NewReport("", rawTestOutput)
		require.NotNil(t, report)
		require.Equal(t, rawTestOutput, report.RawTestOutput)
		require.Empty(t, report.Profiles)
		require.Empty(t, report.CoveredLines)
	})
}

// Tests for [Report.Files] function.
func Test_Report_Files(t *testing.T) {
	t.Run("should return empty slice when no covered lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		files := report.Files()
		require.Empty(t, files)
	})

	t.Run("should return sorted list of file paths", func(t *testing.T) {
		report := NewReport("", []byte{})
		report.Profiles = []*Profile{
			{NoModuleFileName: "b.go"},
			{NoModuleFileName: "a.go"},
		}
		expectedFiles := []string{"a.go", "b.go"}

		files := report.Files()
		require.Equal(t, expectedFiles, files)
	})

	t.Run("should return unique file paths", func(t *testing.T) {
		report := NewReport("", []byte{})
		report.Profiles = []*Profile{
			{NoModuleFileName: "a.go"},
			{NoModuleFileName: "a.go"},
			{NoModuleFileName: "b.go"},
		}
		expectedFiles := []string{"a.go", "b.go"}

		files := report.Files()
		require.Equal(t, expectedFiles, files)
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
	const fileName = "file.go"

	t.Run("single line", func(t *testing.T) {
		report := NewReport("", []byte{})
		block := cover.ProfileBlock{
			StartLine: 1,
			EndLine:   1,
			Count:     1,
		}

		report.expandCoveredLines(fileName, block)

		require.True(t, report.CoveredLines[fileName][1])
	})

	t.Run("multiple lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		block := cover.ProfileBlock{
			StartLine: 1,
			EndLine:   3,
			Count:     1,
		}

		report.expandCoveredLines(fileName, block)

		require.True(t, report.CoveredLines[fileName][1])
		require.True(t, report.CoveredLines[fileName][2])
		require.True(t, report.CoveredLines[fileName][3])
	})

	t.Run("uncovered lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		block := cover.ProfileBlock{
			StartLine: 1,
			EndLine:   3,
			Count:     0,
		}

		report.expandCoveredLines(fileName, block)

		require.False(t, report.CoveredLines[fileName][1])
		require.False(t, report.CoveredLines[fileName][2])
		require.False(t, report.CoveredLines[fileName][3])
	})

	t.Run("covered and uncovered lines", func(t *testing.T) {
		report := NewReport("", []byte{})
		block1 := cover.ProfileBlock{
			StartLine: 1,
			EndLine:   2,
			Count:     1,
		}
		block2 := cover.ProfileBlock{
			StartLine: 3,
			EndLine:   4,
			Count:     0,
		}

		report.expandCoveredLines(fileName, block1)
		report.expandCoveredLines(fileName, block2)

		require.True(t, report.CoveredLines[fileName][1])
		require.True(t, report.CoveredLines[fileName][2])
		require.False(t, report.CoveredLines[fileName][3])
		require.False(t, report.CoveredLines[fileName][4])
	})

	t.Run("covered and uncovered lines in different files", func(t *testing.T) {
		report := NewReport("", []byte{})
		// file1.go
		block1 := cover.ProfileBlock{
			StartLine: 1,
			EndLine:   2,
			Count:     1,
		}
		block2 := cover.ProfileBlock{
			StartLine: 3,
			EndLine:   4,
			Count:     0,
		}

		// file2.go
		block3 := cover.ProfileBlock{
			StartLine: 1,
			EndLine:   2,
			Count:     1,
		}
		block4 := cover.ProfileBlock{
			StartLine: 3,
			EndLine:   4,
			Count:     0,
		}

		report.expandCoveredLines("file1.go", block1)
		report.expandCoveredLines("file1.go", block2)
		report.expandCoveredLines("file2.go", block3)
		report.expandCoveredLines("file2.go", block4)

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
