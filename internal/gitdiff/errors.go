package gitdiff

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/engmtcdrm/uncloak/internal/utils"
)

var (
	ErrGitDoesNotExist = fmt.Errorf("git command not found: ensure git is installed and available in $PATH")
	ErrNotAGitRepo     = fmt.Errorf("not a git repository: ensure you are in a git repository or provide a valid file path")
)

func errNoOutput(cmd *exec.Cmd, targetRef string) error {
	currentBranch := getCurrentBranch()

	var sb bytes.Buffer

	if cmd != nil && len(cmd.Args) > 0 {
		fmt.Fprintf(&sb, "command: %q: ", strings.Join(cmd.Args, " "))
	}

	fmt.Fprint(&sb, "git diff command produced no output.")

	switch {
	case targetRef != "" && currentBranch != "":
		fmt.Fprintf(&sb, " Is the target ref (%s) the same as the current branch (%s)?", targetRef, currentBranch)
	case targetRef != "" && currentBranch == "":
		fmt.Fprintf(&sb, " Is the target ref (%s) the same as the current branch?", targetRef)
	case targetRef == "" && currentBranch != "":
		fmt.Fprintf(&sb, " Is the target ref the same as the current branch (%s)?", currentBranch)
	}

	return errors.New(sb.String())
}

func handleExecError(cmd *exec.Cmd, output []byte, err error) error {
	if execErr, ok := err.(*exec.Error); ok && execErr.Err.Error() == "executable file not found in $PATH" {
		return ErrGitDoesNotExist
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		output = append(output, exitErr.Stderr...)
		return utils.ExecError(cmd, output, err)
	}

	return utils.ExecError(cmd, output, err)
}
