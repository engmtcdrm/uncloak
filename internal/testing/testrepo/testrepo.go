package testrepo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
)

const (
	// NewBranchName is the name of the new development branch created in the
	// test repository.
	NewBranchName = "new-dev"

	cleanupDelay = 200 * time.Millisecond
)

// Init redirects stdout to a temporary file, creates a temporary directory,
// and initializes a git repository in that directory. It returns the path to
// the temporary directory and the file handle for the redirected stdout.
func Init(ctx context.Context, t *testing.T) (string, *os.File) {
	t.Helper()

	stdoutFile := testutils.SetStdout(t)
	tempDir := t.TempDir()
	t.Cleanup(func() {
		// Sleep for a short duration to ensure that any pending operations are completed
		time.Sleep(cleanupDelay)
	})

	// Create a README.md file to ensure the git repository is not empty
	readmeMd := filepath.Join(tempDir, "README.md")
	readmeFile, err := os.Create(readmeMd)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, readmeFile.Close())
	})

	t.Chdir(tempDir)
	testgit.Init(ctx, t)
	testgit.AddCommit(ctx, t, "Initial commit")

	return tempDir, stdoutFile
}

// InitWithFileCopy redirects stdout to a temporary file, creates a temporary
// directory, initializes a git repository in that directory, copies the test
// repo files to the temporary directory, creates a new branch, and commits the
// changes. It returns the path to the temporary directory and the file handle
// for the redirected stdout.
func InitWithFileCopy(ctx context.Context, t *testing.T) (string, *os.File) {
	t.Helper()

	stdoutFile := testutils.SetStdout(t)
	repoPath := testgit.GetTestRepoPath(ctx, t)
	tempDir := t.TempDir()
	t.Cleanup(func() {
		// Sleep for a short duration to ensure that any pending operations are completed
		time.Sleep(cleanupDelay)
	})

	// Create a README.md file to ensure the git repository is not empty
	readmeMd := filepath.Join(tempDir, "README.md")
	readmeFile, err := os.Create(readmeMd)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, readmeFile.Close())
	})

	t.Chdir(tempDir)
	testgit.Init(ctx, t)
	testgit.AddCommit(ctx, t, "Initial commit")

	// Copy the test repo and commit the changes
	testfiles.CopyDir(t, repoPath, tempDir)
	testgit.CreateBranch(ctx, t, NewBranchName)
	testgit.AddCommit(ctx, t, "New development")

	return tempDir, stdoutFile
}
