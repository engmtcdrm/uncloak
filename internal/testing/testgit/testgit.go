package testgit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	MainBranchName = "main"
	TestRepoDir    = "test-repo-covered"
)

var (
	gitInitCmds = []gitCmd{
		{name: "init", args: []string{"git", "init"}},
		{name: "branch", args: []string{"git", "branch", "-m", MainBranchName}},
		{name: "config-user", args: []string{"git", "config", "user.name", "test-user"}},
		{name: "config-email", args: []string{"git", "config", "user.email", "test-user@example.com"}},
	}

	gitAddCommitCmds = []gitCmd{
		{name: "add", args: []string{"git", "add", "."}},
		{name: "commit", args: []string{"git", "commit", "-m"}},
	}

	gitCreateBranchCmds = []gitCmd{
		{name: "branch", args: []string{"git", "branch"}},
		{name: "checkout", args: []string{"git", "checkout"}},
	}
)

type gitCmd struct {
	name string
	args []string
}

// AddCommit adds and commits changes in the git repository with the given
// commit message.
func AddCommit(t *testing.T, commitMsg string) {
	t.Helper()

	repoPath := getwd(t)

	for _, cmd := range gitAddCommitCmds {
		args := cmd.args

		if cmd.name == "commit" {
			args = append(args, commitMsg)
		}

		execCmd(t, args, repoPath)
	}
}

// CreateBranch creates a new branch and checks it out in the git repository.
func CreateBranch(t *testing.T, branchName string) {
	t.Helper()

	repoPath := getwd(t)

	for _, cmd := range gitCreateBranchCmds {
		args := cmd.args
		args = append(args, branchName)
		execCmd(t, args, repoPath)
	}
}

// GetTestRepoPath returns the absolute path to the test git repository
// [TestRepoDir].
func GetTestRepoPath(t *testing.T) string {
	t.Helper()

	repoPath := filepath.Join(rootDir(t), TestRepoDir)
	return repoPath
}

// Init initializes a git repository in the current working directory.
func Init(t *testing.T) {
	t.Helper()

	repoPath := getwd(t)

	for _, cmd := range gitInitCmds {
		execCmd(t, cmd.args, repoPath)
	}

	t.Cleanup(func() {
		err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				os.Chmod(path, 0777)
			}
			return nil
		})
		require.NoError(t, err)
		require.NoError(t, os.RemoveAll(repoPath))
	})
}

// execCmd executes a command in the given repository path and checks for
// errors.
func execCmd(t *testing.T, args []string, repoPath string) {
	t.Helper()

	execCmd := exec.Command(args[0], args[1:]...)
	execCmd.Dir = repoPath
	output, err := execCmd.CombinedOutput()
	require.NoError(t, err, "Failed to run command %q, output: %s", args, string(output))
}

// getwd returns the current working directory.
func getwd(t *testing.T) string {
	t.Helper()

	repoPath, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	return repoPath
}

// rootDir returns the root directory of the git repository.
func rootDir(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := cmd.Output()
	require.NoError(t, err, "Failed to get git root directory")

	rootPath := strings.TrimSpace(string(output))
	rootPath, err = filepath.EvalSymlinks(rootPath)
	require.NoError(t, err, "Failed to resolve git root directory")

	return rootPath
}
