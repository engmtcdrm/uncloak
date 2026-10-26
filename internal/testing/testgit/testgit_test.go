package testgit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [AddCommit] function.
func Test_AddCommit(t *testing.T) {
	t.Run("should add and commit changes in a git repository in the test repo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		// Create a README.md file to ensure the git repository is not empty
		readmeMd := filepath.Join(tempDir, "README.md")
		readmeFile, err := os.Create(readmeMd)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, readmeFile.Close())
		})

		Init(t.Context(), t)
		AddCommit(t.Context(), t, "Initial commit")
	})
}

// Tests for [Checkout] function.
func Test_Checkout(t *testing.T) {
	t.Run("should checkout a branch in a git repository in the test repo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		const testBranch = "test-branch"

		// Create a README.md file to ensure the git repository is not empty
		readmeMd := filepath.Join(tempDir, "README.md")
		readmeFile, err := os.Create(readmeMd)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, readmeFile.Close())
		})

		Init(t.Context(), t)
		AddCommit(t.Context(), t, "Initial commit")
		CreateBranch(t.Context(), t, testBranch)
		Checkout(t.Context(), t, testBranch)

		currentBranchName := RunGit(t.Context(), t, []string{"branch", "--show-current"}, tempDir)
		require.Contains(t, string(currentBranchName), testBranch)
	})
}

// Tests for [CreateBranch] function.
func Test_CreateBranch(t *testing.T) {
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

		Init(t.Context(), t)
		AddCommit(t.Context(), t, "Initial commit")
		CreateBranch(t.Context(), t, "test-branch")
	})
}

// Tests for [GetTestRepoPath] function.
func Test_GetTestRepoPath(t *testing.T) {
	t.Run("should return the absolute path to the test git repository", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		Init(t.Context(), t)

		expectedRootPath, err := filepath.EvalSymlinks(tempDir)
		require.NoError(t, err)

		expectedPath := filepath.Join(expectedRootPath, TestRepoDir)
		result := GetTestRepoPath(t.Context(), t)
		require.Equal(t, expectedPath, result)
	})
}

// Tests for [Init] function.
func Test_Init(t *testing.T) {
	t.Run("should initialize a git repository in the test repo directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		Init(t.Context(), t)
	})
}

// Tests for [RootDir] function.
func Test_RootDir(t *testing.T) {
	t.Run("should return the root directory of the temp directory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)
		Init(t.Context(), t)

		subDir := filepath.Join(tempDir, "subdir")
		err := os.Mkdir(subDir, os.ModePerm)
		require.NoError(t, err)

		t.Chdir(subDir)
		result := RootDir(t.Context(), t)
		expectedRootPath, err := filepath.EvalSymlinks(tempDir)
		require.NoError(t, err)

		require.Equal(t, expectedRootPath, result)
	})
}

// Tests for [RunGit] function.
func Test_RunGit(t *testing.T) {
	t.Run("should execute a command and return its output", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)
		output := RunGit(t.Context(), t, []string{"version"}, tempDir)
		require.Contains(t, string(output), "git version")
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
