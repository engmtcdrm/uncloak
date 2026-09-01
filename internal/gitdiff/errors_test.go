package gitdiff

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [NewInvalidRefError] function.
func Test_NewInvalidRefError(t *testing.T) {
	expectedRef := "invalid-ref"
	expectedTargetRef := true

	t.Run("should create a new InvalidRefError with the given ref and targetRef", func(t *testing.T) {
		errInvalidRef := &InvalidRefError{}

		err := NewInvalidRefError(expectedRef, expectedTargetRef)
		require.Error(t, err)
		require.ErrorAs(t, err, &errInvalidRef)
		assert.Equal(t, expectedRef, err.ref)
		assert.Equal(t, expectedTargetRef, err.targetRef)
	})
}

// Tests for [InvalidRefError.Error] function.
func Test_InvalidRefError_Error(t *testing.T) {
	t.Run("should return the error message for target reference", func(t *testing.T) {
		errInvalidRef := &InvalidRefError{}
		err := NewInvalidRefError("invalid-ref", true)

		require.Error(t, err)
		require.ErrorAs(t, err, &errInvalidRef)

		expectedMessage := errPrefix + " invalid target reference: invalid-ref"
		assert.Equal(t, expectedMessage, err.Error())
	})

	t.Run("should return the error message for non-target reference", func(t *testing.T) {
		errInvalidRef := &InvalidRefError{}
		err := NewInvalidRefError("invalid-ref", false)

		require.Error(t, err)
		require.ErrorAs(t, err, &errInvalidRef)

		expectedMessage := errPrefix + " invalid reference: invalid-ref"
		assert.Equal(t, expectedMessage, err.Error())
	})
}

// Tests for [NewSameRefError] function.
func Test_NewSameRefError(t *testing.T) {
	expectedTargetRef := "target-ref"
	expectedCurrentBranch := "current-branch"

	t.Run("should create a new SameRefError with the given target reference and current HEAD ref", func(t *testing.T) {
		errSameRef := &SameRefError{}

		err := NewSameRefError(expectedTargetRef, expectedCurrentBranch)
		require.Error(t, err)
		require.ErrorAs(t, err, &errSameRef)
		assert.Equal(t, expectedTargetRef, err.targetRef)
		assert.Equal(t, expectedCurrentBranch, err.currentHeadRef)
	})
}

// Tests for [SameRefError.Error] function.
func Test_SameRefError_Error(t *testing.T) {
	t.Run("should return the error message", func(t *testing.T) {
		errSameRef := &SameRefError{}
		err := NewSameRefError("target-ref", "current-branch")

		require.Error(t, err)
		require.ErrorAs(t, err, &errSameRef)

		expectedMessage := errPrefix + " target reference (target-ref) is the same as the current HEAD (current-branch)"
		assert.Equal(t, expectedMessage, err.Error())
	})
}

// Tests for [handleExecError] function.
func Test_handleExecError(t *testing.T) {
	ctx := context.Background()

	t.Run("should return ErrGitDoesNotExist when git is not found", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("PATH", tempDir)
		cmd := exec.CommandContext(ctx, "git", "diff")
		err := handleExecError(cmd, nil, &exec.Error{Name: "git", Err: exec.ErrNotFound})
		require.ErrorIs(t, err, exec.ErrNotFound)
	})

	t.Run("should return an error with command and output when exec fails", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, "git", "diff")
		err := handleExecError(cmd, []byte("output"), &exec.ExitError{Stderr: []byte("stderr")})
		require.Error(t, err)
		require.Contains(t, err.Error(), "git diff")
		require.Contains(t, err.Error(), "output")
		require.Contains(t, err.Error(), "stderr")
	})

	t.Run("should return an error with command and output when exec fails with unknown error", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, "git", "diff")
		err := handleExecError(cmd, []byte("output"), errors.New("unknown error"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "git diff")
		require.Contains(t, err.Error(), "output")
	})
}
