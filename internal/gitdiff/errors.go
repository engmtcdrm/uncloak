package gitdiff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/utils"
)

var (
	// ErrNotAGitRepo indicates that the current directory is not a git
	// repository.
	ErrNotAGitRepo = errors.New("not a git repository: ensure you are in a git repository")

	// ErrNoCurrentBranch indicates that there is no current branch found in the
	// git repository. This can happen if the repository is in a detached HEAD
	// state or if there are no commits in the repository.
	ErrNoCurrentBranch = errors.New("no current branch found: ensure you are on a valid git branch. If you are in a detached HEAD state, please provide a valid target ref to compare against")
)

type ErrSameBranch struct {
	TargetRef     string
	CurrentBranch string
}

func NewErrSameBranch(targetRef, currentBranch string) *ErrSameBranch {
	return &ErrSameBranch{
		TargetRef:     targetRef,
		CurrentBranch: currentBranch,
	}
}

func (e *ErrSameBranch) Error() string {
	return fmt.Sprintf("target ref (%s) is the same as the current branch (%s)", pp.Red(e.TargetRef), pp.Red(e.CurrentBranch))
}

func errNoOutput(ctx context.Context, cmd *exec.Cmd, targetRef string) error {
	currentBranch := getCurrentBranch(ctx)

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
