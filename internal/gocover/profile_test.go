package gocover

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [NewProfile] function.
func Test_NewProfile(t *testing.T) {
	t.Run("return empty profile with empty rawTestOutput", func(t *testing.T) {
		profile := NewProfile("", []byte{})
		require.NotNil(t, profile)
		require.Empty(t, profile.RawTestOutput)
		require.Empty(t, profile.Lines)
		require.Empty(t, profile.CoveredLines)
	})

	t.Run("return empty profile with nil rawTestOutput", func(t *testing.T) {
		profile := NewProfile("", nil)
		require.NotNil(t, profile)
		require.Empty(t, profile.RawTestOutput)
		require.Empty(t, profile.Lines)
		require.Empty(t, profile.CoveredLines)
	})

	t.Run("return profile with rawTestOutput", func(t *testing.T) {
		rawTestOutput := []byte("test output")
		profile := NewProfile("", rawTestOutput)
		require.NotNil(t, profile)
		require.Equal(t, rawTestOutput, profile.RawTestOutput)
		require.Empty(t, profile.Lines)
		require.Empty(t, profile.CoveredLines)
	})
}

// Tests for [Profile.Files] function.
func Test_Profile_Files(t *testing.T) {
	t.Run("return empty slice when no covered lines", func(t *testing.T) {
		profile := NewProfile("", []byte{})
		files := profile.Files()
		require.Empty(t, files)
	})

	t.Run("return sorted list of file paths", func(t *testing.T) {
		profile := NewProfile("", []byte{})
		profile.CoveredLines["b.go"] = map[int]bool{1: true}
		profile.CoveredLines["a.go"] = map[int]bool{1: true}

		files := profile.Files()
		require.Equal(t, []string{"a.go", "b.go"}, files)
	})
}

// Tests for [Profile.IsInTestCoverage] function.
func Test_Profile_IsInTestCoverage(t *testing.T) {
	profile := NewProfile("", []byte{})

	profile.CoveredLines["file.go"] = map[int]bool{
		1: true,
		3: true,
	}

	require.True(t, profile.IsInTestCoverage("file.go", 1))
	require.False(t, profile.IsInTestCoverage("file.go", 2))
	require.False(t, profile.IsInTestCoverage("other.go", 1))
}

// Tests for [Profile.expandCoveredLines] function.
func Test_Profile_expandCoveredLines(t *testing.T) {
	profile := NewProfile("", []byte{})

	t.Run("single line", func(t *testing.T) {
		line := Line{
			Path:      "file.go",
			StartLine: 1,
			EndLine:   1,
			Count:     1,
		}
		profile.expandCoveredLines(line)

		require.True(t, profile.CoveredLines["file.go"][1])
	})

	t.Run("multiple lines", func(t *testing.T) {
		line := Line{
			Path:      "file.go",
			StartLine: 1,
			EndLine:   3,
			Count:     1,
		}
		profile.expandCoveredLines(line)

		require.True(t, profile.CoveredLines["file.go"][1])
		require.True(t, profile.CoveredLines["file.go"][2])
		require.True(t, profile.CoveredLines["file.go"][3])
	})

	t.Run("uncovered lines", func(t *testing.T) {
		line := Line{
			Path:      "file.go",
			StartLine: 1,
			EndLine:   3,
			Count:     0,
		}
		profile.expandCoveredLines(line)

		require.False(t, profile.CoveredLines["file.go"][1])
		require.False(t, profile.CoveredLines["file.go"][2])
		require.False(t, profile.CoveredLines["file.go"][3])
	})

	t.Run("covered and uncovered lines", func(t *testing.T) {
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
		profile.expandCoveredLines(line1)
		profile.expandCoveredLines(line2)

		require.True(t, profile.CoveredLines["file1.go"][1])
		require.True(t, profile.CoveredLines["file1.go"][2])
		require.False(t, profile.CoveredLines["file1.go"][3])
		require.False(t, profile.CoveredLines["file1.go"][4])
	})

	t.Run("covered and uncovered lines in different files", func(t *testing.T) {
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
		profile.expandCoveredLines(line1)
		profile.expandCoveredLines(line2)
		profile.expandCoveredLines(line3)
		profile.expandCoveredLines(line4)

		require.True(t, profile.CoveredLines["file1.go"][1])
		require.True(t, profile.CoveredLines["file1.go"][2])
		require.False(t, profile.CoveredLines["file1.go"][3])
		require.False(t, profile.CoveredLines["file1.go"][4])
		require.True(t, profile.CoveredLines["file2.go"][1])
		require.True(t, profile.CoveredLines["file2.go"][2])
		require.False(t, profile.CoveredLines["file2.go"][3])
		require.False(t, profile.CoveredLines["file2.go"][4])
	})
}
