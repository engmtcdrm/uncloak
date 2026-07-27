package testrepo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
)

const NewBranchName = "new-dev"

// Init redirects stdout to a temporary file, creates a temporary directory,
// and initializes a git repository in that directory. It returns the path to
// the temporary directory and the file handle for the redirected stdout.
func Init(t *testing.T) (string, *os.File) {
	t.Helper()

	stdoutFile := testutils.SetStdout(t)
	tempDir := t.TempDir()
	t.Cleanup(func() {
		_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				_ = os.Chmod(path, 0777)
			}
			return nil
		})
	})

	// Create a README.md file to ensure the git repository is not empty
	readmeMd := filepath.Join(tempDir, "README.md")
	readmeFile, err := os.Create(readmeMd)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, readmeFile.Close())
	})

	t.Chdir(tempDir)
	testgit.Init(t)
	testgit.AddCommit(t, "Initial commit")

	return tempDir, stdoutFile
}

// InitWithFileCopy redirects stdout to a temporary file, creates a temporary
// directory, initializes a git repository in that directory, copies the test
// repo files to the temporary directory, creates a new branch, and commits the
// changes. It returns the path to the temporary directory and the file handle
// for the redirected stdout.
func InitWithFileCopy(t *testing.T) (string, *os.File) {
	t.Helper()

	stdoutFile := testutils.SetStdout(t)
	repoPath := testgit.GetTestRepoPath(t)
	tempDir := t.TempDir()
	t.Cleanup(func() {
		_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				_ = os.Chmod(path, 0777)
			}
			return nil
		})
	})

	// Create a README.md file to ensure the git repository is not empty
	readmeMd := filepath.Join(tempDir, "README.md")
	readmeFile, err := os.Create(readmeMd)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, readmeFile.Close())
	})

	t.Chdir(tempDir)
	testgit.Init(t)
	testgit.AddCommit(t, "Initial commit")

	// Copy the test repo and commit the changes
	testfiles.CopyDir(t, repoPath, tempDir)
	testgit.CreateBranch(t, NewBranchName)
	testgit.AddCommit(t, "New development")

	return tempDir, stdoutFile
}
