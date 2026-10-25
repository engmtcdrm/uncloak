package testgit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// MainBranchName represents the name of the main branch in the git
	// repository.
	MainBranchName = "main"

	// TestRepoDir represents the directory name of the test git repository.
	TestRepoDir = "test-repo-covered"
)

var (
	gitInitCmds = []gitArgs{
		{name: "init", args: []string{"init"}},
		{name: "branch", args: []string{"branch", "-m", MainBranchName}},
		{name: "config-user", args: []string{"config", "user.name", "test-user"}},
		{name: "config-email", args: []string{"config", "user.email", "test-user@example.com"}},
	}

	gitAddCommitCmds = []gitArgs{
		{name: "add", args: []string{"add", "."}},
		{name: "commit", args: []string{"commit", "-m"}},
	}
)

type gitArgs struct {
	name string
	args []string
}

// AddCommit adds and commits changes in the git repository with the given
// commit message.
func AddCommit(ctx context.Context, t *testing.T, commitMsg string) {
	t.Helper()

	repoPath := getwd(t)

	for _, cmd := range gitAddCommitCmds {
		args := cmd.args

		if cmd.name == "commit" {
			args = append(args, commitMsg)
		}

		_ = RunGit(ctx, t, args, repoPath)
	}
}

// Checkout checks out the specified branch in the git repository.
func Checkout(ctx context.Context, t *testing.T, branchName string) {
	t.Helper()

	repoPath := getwd(t)
	_ = RunGit(ctx, t, []string{"checkout", branchName}, repoPath)
}

// CreateBranch creates a new branch and checks it out in the git repository.
func CreateBranch(ctx context.Context, t *testing.T, branchName string) {
	t.Helper()

	repoPath := getwd(t)

	_ = RunGit(ctx, t, []string{"branch", branchName}, repoPath)
}

// GetTestRepoPath returns the absolute path to the test git repository
// [TestRepoDir].
func GetTestRepoPath(ctx context.Context, t *testing.T) string {
	t.Helper()

	repoPath := filepath.Join(RootDir(ctx, t), TestRepoDir)
	return repoPath
}

// Init initializes a git repository in the current working directory.
func Init(ctx context.Context, t *testing.T) {
	t.Helper()

	repoPath := getwd(t)

	for _, cmd := range gitInitCmds {
		_ = RunGit(ctx, t, cmd.args, repoPath)
	}
}

// RootDir returns the root directory of the git repository.
func RootDir(ctx context.Context, t *testing.T) string {
	t.Helper()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")

	output, err := cmd.Output()
	require.NoError(t, err, "Failed to get git root directory")

	rootPath := strings.TrimSpace(string(output))
	rootPath, err = filepath.EvalSymlinks(rootPath)
	require.NoError(t, err, "Failed to resolve git root directory")

	return rootPath
}

// RunGit executes a command in the given repository path and checks for errors.
func RunGit(ctx context.Context, t *testing.T, args []string, repoPath string) []byte {
	t.Helper()

	execCmd := exec.CommandContext(ctx, "git", args...)
	execCmd.Dir = repoPath
	output, err := execCmd.CombinedOutput()
	require.NoError(t, err, "Failed to run command %q, output: %s", args, string(output))

	return output
}

// getwd returns the current working directory.
func getwd(t *testing.T) string {
	t.Helper()

	repoPath, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	return repoPath
}
