package gitdiff

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/engmtcdrm/uncloak/internal/utils"
)

const (
	errPrefix = "git diff:"
)

var (
	// ErrNoOutput indicates that the git diff command produced no output.
	ErrNoOutput = errors.New(errPrefix + " produced no output")

	// ErrNotAGitRepo indicates that the current directory is not a git
	// repository.
	ErrNotAGitRepo = errors.New(errPrefix + " not a git repository: ensure you are in a git repository")
)

// InvalidRefError indicates that the provided reference is invalid.
type InvalidRefError struct {
	ref       string
	targetRef bool
}

// NewInvalidRefError creates a new [InvalidRefError] with the given reference.
func NewInvalidRefError(ref string, targetRef bool) *InvalidRefError {
	return &InvalidRefError{
		ref:       ref,
		targetRef: targetRef,
	}
}

// Error returns the error message.
func (e *InvalidRefError) Error() string {
	if e.targetRef {
		return fmt.Sprintf("%s invalid target reference: %s", errPrefix, e.ref)
	}

	return fmt.Sprintf("%s invalid reference: %s", errPrefix, e.ref)
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
