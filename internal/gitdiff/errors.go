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
	// ErrNotAGitRepo indicates that the current directory is not a git
	// repository.
	ErrNotAGitRepo = errors.New("not a git repository: ensure you are in a git repository or provide a valid file path")

	// ErrNoParentBranch indicates that no parent branch could be found, which
	// may occur if the current branch is the main branch.
	ErrNoParentBranch = errors.New("no parent branch found: are you on the main branch for this repository?")
)

func errNoOutput(cmd *exec.Cmd, targetRef string) error {
	currentBranch := getCurrentBranch()

	var buf bytes.Buffer

	if cmd != nil && len(cmd.Args) > 0 {
		fmt.Fprintf(&buf, "command: %q: ", strings.Join(cmd.Args, " "))
	}

	fmt.Fprint(&buf, "git diff command produced no output.")

	switch {
	case targetRef != "" && currentBranch != "":
		fmt.Fprintf(&buf, " Is the target ref (%s) the same as the current branch (%s)?", targetRef, currentBranch)
	case targetRef != "" && currentBranch == "":
		fmt.Fprintf(&buf, " Is the target ref (%s) the same as the current branch?", targetRef)
	case targetRef == "" && currentBranch != "":
		fmt.Fprintf(&buf, " Is the target ref the same as the current branch (%s)?", currentBranch)
	}

	return errors.New(buf.String())
}

func handleExecError(cmd *exec.Cmd, output []byte, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return err
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		output = append(output, exitErr.Stderr...)
		return utils.ExecError(cmd, output, err)
	}

	return utils.ExecError(cmd, output, err)
}
