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

// SameRefError indicates that the target reference is the same as the current
// HEAD reference.
type SameRefError struct {
	targetRef      string
	currentHeadRef string
}

// NewSameRefError creates a new [SameRefError] with the given target reference
// and current HEAD reference.
func NewSameRefError(targetRef, currentHeadRef string) *SameRefError {
	return &SameRefError{
		targetRef:      targetRef,
		currentHeadRef: currentHeadRef,
	}
}

// Error returns the error message.
func (e *SameRefError) Error() string {
	return fmt.Sprintf("%s target reference (%s) is the same as the current HEAD reference (%s)", errPrefix, e.targetRef, e.currentHeadRef)
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
