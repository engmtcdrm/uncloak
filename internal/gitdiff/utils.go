package gitdiff

import (
	"context"
	"os/exec"
	"strings"
)

const (
	headRef     = "HEAD"
	gitCmd      = "git"
	revParseCmd = "rev-parse"
)

// gitRootDir retrieves the root directory of the current Git repository by
// executing the "git rev-parse --show-toplevel" command. It returns the root
// directory path if the command is successful, or an error if there is an
// issue executing the command.
func gitRootDir(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, gitCmd, revParseCmd, "--show-toplevel")

	output, err := cmd.Output()
	if err != nil {
		return "", handleExecError(cmd, output, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// isGitDir checks if the current working directory is a Git repository by
// executing the "git rev-parse --git-dir" command. if it returns true it means
// the current working directory is in a git repository. If it returns false, it
// means the current working directory is not in a git repository.
func isGitDir(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, gitCmd, revParseCmd, "--git-dir")

	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(out) > 0
}

// isGoFile checks if the given file path corresponds to a Go source file (i.e.,
// it ends with ".go" but not with "_test.go").
func isGoFile(filePath string) bool {
	return strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go")
}

// validateRef retrieves the commit hash of the provided Git reference by
// executing the `git rev-parse --verify '<ref>^{commit}'` command. It returns a
// boolean indicating whether the reference is valid or not.
func validateRef(ctx context.Context, ref string) bool {
	peelRef := ref + "^{commit}"

	cmd := exec.CommandContext(ctx, gitCmd, revParseCmd, "--verify", peelRef)
	err := cmd.Run()

	return err == nil
}
