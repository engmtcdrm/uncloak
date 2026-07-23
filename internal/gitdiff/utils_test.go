package gitdiff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [gitRootDir] function.
func Test_gitRootDir(t *testing.T) {
	t.Run("should return the root directory of this repo", func(t *testing.T) {
		rootDir, err := gitRootDir()
		require.NoError(t, err)
		require.NotEmpty(t, rootDir)
	})

	t.Run("should return an error when exec fails", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("PATH", tempDir)
		_, err := gitRootDir()
		require.Error(t, err)
	})
}

// Tests for [isGitDir] function.
func Test_isGitDir(t *testing.T) {
	t.Run("should return true with this repo", func(t *testing.T) {
		require.True(t, isGitDir())
	})

	t.Run("should return false with a non-git directory", func(t *testing.T) {
		t.Chdir("/tmp")
		require.False(t, isGitDir())
	})
}

// Tests for [isGoFile] function.
func Test_isGoFile(t *testing.T) {
	t.Run("returns true for .go files", func(t *testing.T) {
		require.True(t, isGoFile("file.go"))
	})

	t.Run("returns false for _test.go files", func(t *testing.T) {
		require.False(t, isGoFile("file_test.go"))
	})

	t.Run("returns false for non-.go files", func(t *testing.T) {
		require.False(t, isGoFile("file.txt"))
	})
}
