package gitdiff

import (
	"context"
	"os"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [getRef] function.
func Test_getRef(t *testing.T) {
	ctx := context.Background()

	t.Run("should return commit hash and true for valid reference", func(t *testing.T) {
		testrepo.InitWithFileCopy(ctx, t)

		commitHash, isValid := getRef(ctx, LocalMain)
		require.True(t, isValid)
		require.NotEmpty(t, commitHash)
	})

	t.Run("should return commit hash and true for HEAD reference", func(t *testing.T) {
		testrepo.InitWithFileCopy(ctx, t)

		commitHash, isValid := getRef(ctx, "HEAD")
		require.True(t, isValid)
		require.NotEmpty(t, commitHash)
	})

	t.Run("should return empty commit hash and false for invalid reference", func(t *testing.T) {
		testrepo.InitWithFileCopy(ctx, t)

		commitHash, isValid := getRef(ctx, "invalid-ref")
		require.False(t, isValid)
		require.Empty(t, commitHash)
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
