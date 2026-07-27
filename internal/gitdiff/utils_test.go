package gitdiff

import (
	"os"
	"os/exec"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [findNearestParent] function.
func Test_findNearestParent(t *testing.T) {
	t.Run("should return empty parent if on main branch", func(t *testing.T) {
		_, _ = testrepo.Init(t)
		parent := findNearestParent()
		require.Equal(t, "", parent)
	})

	t.Run("should return empty if directory is not a git repo", func(t *testing.T) {
		t.Chdir(t.TempDir())
		parent := findNearestParent()
		require.Equal(t, "", parent)
	})

	t.Run("should return parent branch if on a child branch", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		parent := findNearestParent()
		require.Equal(t, LocalMain, parent)
	})

	t.Run("should return empty if git is in a detached HEAD state", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		cmd := exec.Command("git", "checkout", "--detach", "HEAD")
		require.NoError(t, cmd.Run())

		parent := findNearestParent()
		require.Equal(t, "", parent)
	})

	t.Run("should return empty if git repo has little to no commits", func(t *testing.T) {
		_, _ = testrepo.Init(t)
		parent := findNearestParent()
		require.Equal(t, "", parent)
	})
}

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
		t.Chdir(os.TempDir())
		require.False(t, isGitDir())
	})
}

// Tests for [hasParent] function.
func Test_hasParent(t *testing.T) {
	t.Run("should return false with no parent", func(t *testing.T) {
		_, _ = testrepo.Init(t)
		require.False(t, hasParent())
	})

	t.Run("should return true with a parent", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		require.True(t, hasParent())
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
