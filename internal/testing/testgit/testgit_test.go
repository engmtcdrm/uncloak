package testgit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [AddCommit] function.
func Test_AddCommit(t *testing.T) {
	t.Run("should add and commit changes in a git repository in the test repo directory", func(t *testing.T) {
		ctx := context.Background()
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		// Create a README.md file to ensure the git repository is not empty
		readmeMd := filepath.Join(tempDir, "README.md")
		readmeFile, err := os.Create(readmeMd)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, readmeFile.Close())
		})

		Init(ctx, t)
		AddCommit(ctx, t, "Initial commit")
	})
}

// Tests for [CreateBranch] function.
func Test_CreateBranch(t *testing.T) {
	ctx := context.Background()

	t.Run("should create a branch in a git repository in the test repo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		// Create a README.md file to ensure the git repository is not empty
		readmeMd := filepath.Join(tempDir, "README.md")
		readmeFile, err := os.Create(readmeMd)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, readmeFile.Close())
		})

		Init(ctx, t)
		AddCommit(ctx, t, "Initial commit")
		CreateBranch(ctx, t, "test-branch")
	})
}

// Tests for [GetTestRepoPath] function.
func Test_GetTestRepoPath(t *testing.T) {
	ctx := context.Background()

	t.Run("should return the absolute path to the test git repository", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		Init(ctx, t)

		expectedRootPath, err := filepath.EvalSymlinks(tempDir)
		require.NoError(t, err)

		expectedPath := filepath.Join(expectedRootPath, TestRepoDir)
		result := GetTestRepoPath(ctx, t)
		require.Equal(t, expectedPath, result)
	})
}

// Tests for [Init] function.
func Test_Init(t *testing.T) {
	ctx := context.Background()

	t.Run("should initialize a git repository in the test repo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		Init(ctx, t)
	})
}

// Tests for [execCmd] function.
func Test_execCmd(t *testing.T) {
	t.Run("should execute a command and return its output", func(t *testing.T) {
		ctx := context.Background()
		tempDir := t.TempDir()
		t.Chdir(tempDir)
		execCmd(ctx, t, []string{"git", "version"}, tempDir)
	})
}

// Tests for [getwd] function.
func Test_getwd(t *testing.T) {
	t.Run("should return the current working directory", func(t *testing.T) {
		expectedWd, err := os.Getwd()
		require.NoError(t, err)

		require.Equal(t, expectedWd, getwd(t))
	})
}

// Tests for [rootDir] function.
func Test_rootDir(t *testing.T) {
	ctx := context.Background()

	t.Run("should return the root directory of the temp directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)
		Init(ctx, t)

		subDir := filepath.Join(tempDir, "subdir")
		err := os.Mkdir(subDir, os.ModePerm)
		require.NoError(t, err)

		t.Chdir(subDir)
		result := rootDir(ctx, t)
		expectedRootPath, err := filepath.EvalSymlinks(tempDir)
		require.NoError(t, err)

		require.Equal(t, expectedRootPath, result)
	})
}
