package gitdiff

import (
	"os/exec"
	"regexp"
	"strings"
)

const (
	branchBracketPattern      = `.*\[(.*)\].*`
	branchSpecialCharsPattern = `[\^~].*`
)

var (
	regexBranchBracket      = regexp.MustCompile(branchBracketPattern)
	regexBranchSpecialChars = regexp.MustCompile(branchSpecialCharsPattern)
)

// findNearestParent retrieves the name of the nearest parent branch of the
// current branch by executing the "git show-branch" command and parsing its
// output. It returns the name of the nearest parent branch if found or an empty
// string otherwise.
func findNearestParent() string {
	cmd := exec.Command("git", "show-branch")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	currentBranch := getCurrentBranch()
	if currentBranch == "" {
		return ""
	}

	var firstParentLine string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		switch {
		// Skip lines containing the current branch name. We only care about
		// other branches than the current one.
		case strings.Contains(line, currentBranch):
			continue
		// Skip any lines without a "*". The "*" indicates the line is part of
		// the current branch's ancestry.
		case !strings.Contains(line, "*"):
			continue
		}

		firstParentLine = line
		break
	}

	if firstParentLine == "" {
		return ""
	}

	groupMatches := regexBranchBracket.FindStringSubmatch(firstParentLine)
	if len(groupMatches) < 2 {
		return ""
	}

	return regexBranchSpecialChars.ReplaceAllString(groupMatches[1], "")
}

// getCurrentBranch retrieves the name of the current Git branch by executing
// the "git branch --show-current" command. It returns the branch name if the
// command is successful, or an empty string if there is an error.
func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// gitRootDir retrieves the root directory of the current Git repository by
// executing the "git rev-parse --show-toplevel" command. It returns the root
// directory path if the command is successful, or an error if there is an
// issue executing the command.
func gitRootDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := cmd.Output()
	if err != nil {
		return "", handleExecError(cmd, output, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// hasParent checks if the current Git branch has a parent branch by calling
// the findNearestParent function. It returns true if a parent branch is found,
// and false otherwise.
func hasParent() bool {
	return findNearestParent() != ""
}

// isGitDir checks if the current working directory is a git repository by
// executing the "git rev-parse --git-dir" command. if it returns true it means
// the current working directory is in a git repository. If it returns false, it
// means the current working directory is not in a git repository.
func isGitDir() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")

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
