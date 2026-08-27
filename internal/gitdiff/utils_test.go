package gitdiff

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [findNearestParent] function.
func Test_findNearestParent(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty parent if on main branch", func(t *testing.T) {
		_, _ = testrepo.Init(ctx, t)
		parent := findNearestParent(ctx)
		require.Empty(t, parent)
	})

	t.Run("should return empty if directory is not a git repo", func(t *testing.T) {
		t.Chdir(t.TempDir())
		parent := findNearestParent(ctx)
		require.Empty(t, parent)
	})

	t.Run("should return parent branch if on a child branch", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		parent := findNearestParent(ctx)
		require.Equal(t, LocalMain, parent)
	})

	t.Run("should return empty if git is in a detached HEAD state", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		cmd := exec.CommandContext(ctx, "git", "checkout", "--detach", "HEAD")
		require.NoError(t, cmd.Run())

		parent := findNearestParent(ctx)
		require.Empty(t, parent)
	})

	t.Run("should return empty if git repo has little to no commits", func(t *testing.T) {
		_, _ = testrepo.Init(ctx, t)
		parent := findNearestParent(ctx)
		require.Empty(t, parent)
	})
}

// Tests for [gitRootDir] function.
func Test_gitRootDir(t *testing.T) {
	ctx := context.Background()

	t.Run("should return the root directory of this repo", func(t *testing.T) {
		rootDir, err := gitRootDir(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, rootDir)
	})

	t.Run("should return an error when exec fails", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("PATH", tempDir)
		_, err := gitRootDir(ctx)
		require.Error(t, err)
	})
}

// Tests for [isGitDir] function.
func Test_isGitDir(t *testing.T) {
	ctx := context.Background()

	t.Run("should return true with this repo", func(t *testing.T) {
		require.True(t, isGitDir(ctx))
	})

	t.Run("should return false with a non-git directory", func(t *testing.T) {
		t.Chdir(os.TempDir())
		require.False(t, isGitDir(ctx))
	})
}

// Tests for [hasParent] function.
func Test_hasParent(t *testing.T) {
	ctx := context.Background()

	t.Run("should return false with no parent", func(t *testing.T) {
		_, _ = testrepo.Init(ctx, t)
		require.False(t, hasParent(ctx))
	})

	t.Run("should return true with a parent", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		require.True(t, hasParent(ctx))
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
