package gitdiff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [NewResults] function.
func Test_NewResults(t *testing.T) {
	t.Run("should return a Results instance with the root directory", func(t *testing.T) {
		results, err := NewResults()
		require.NoError(t, err)
		require.NotNil(t, results)
		require.NotEmpty(t, results.RootDir)
		require.Empty(t, results.NewLines)
	})

	t.Run("should return an error when exec fails", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("PATH", tempDir)
		results, err := NewResults()
		require.Error(t, err)
		require.Nil(t, results)
	})
}

// Tests for [Results.Files] method.
func Test_Results_Files(t *testing.T) {
	t.Run("return empty slice when no new lines", func(t *testing.T) {
		results, err := NewResults()
		require.NoError(t, err)

		files := results.Files()
		require.Empty(t, files)
	})

	t.Run("return sorted list of file paths", func(t *testing.T) {
		results, err := NewResults()
		require.NoError(t, err)
		results.NewLines["b.go"] = map[int]bool{1: true}
		results.NewLines["a.go"] = map[int]bool{1: true}

		files := results.Files()
		require.Equal(t, []string{"a.go", "b.go"}, files)
	})
}
