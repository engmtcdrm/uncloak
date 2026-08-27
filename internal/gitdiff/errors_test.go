package gitdiff

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [errNoOutput] function.
func Test_errNoOutput(t *testing.T) {
	ctx := context.Background()

	t.Run("nil input produces error with empty command", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		expectedErr := fmt.Sprintf("git diff command produced no output. Is the target ref (%s) the same as the current branch (%s)?",
			testgit.MainBranchName,
			testrepo.NewBranchName,
		)

		err := errNoOutput(ctx, nil, testgit.MainBranchName)
		require.Error(t, err)
		require.Equal(t, expectedErr, err.Error())
	})

	t.Run("non-nil input produces error with command", func(t *testing.T) {
		_, _ = testrepo.Init(ctx, t)
		expectedErr := fmt.Sprintf("command: \"git diff --cached\": git diff command produced no output. Is the target ref (%s) the same as the current branch (%s)?",
			testgit.MainBranchName,
			testgit.MainBranchName,
		)

		cmd := exec.CommandContext(ctx, "git", "diff", "--cached")
		err := errNoOutput(ctx, cmd, testgit.MainBranchName)
		require.Error(t, err)
		require.Equal(t, expectedErr, err.Error())
	})

	t.Run("empty targetRef produces error without targetRef", func(t *testing.T) {
		_, _ = testrepo.Init(ctx, t)
		expectedErr := fmt.Sprintf("command: \"git diff --cached\": git diff command produced no output. Is the target ref the same as the current branch (%s)?",
			testgit.MainBranchName,
		)

		cmd := exec.CommandContext(ctx, "git", "diff", "--cached")
		err := errNoOutput(ctx, cmd, "")
		require.Error(t, err)
		require.Equal(t, expectedErr, err.Error())
	})

	t.Run("empty currentBranch produces error without currentBranch", func(t *testing.T) {
		expectedErr := fmt.Sprintf("command: \"git diff --cached\": git diff command produced no output. Is the target ref (%s) the same as the current branch?",
			testgit.MainBranchName,
		)

		t.Chdir(t.TempDir())

		cmd := exec.CommandContext(ctx, "git", "diff", "--cached")
		err := errNoOutput(ctx, cmd, testgit.MainBranchName)
		require.Error(t, err)
		require.Equal(t, expectedErr, err.Error())
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
