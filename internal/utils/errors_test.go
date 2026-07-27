package utils

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [ExecError] function.
func Test_ExecError(t *testing.T) {
	t.Run("ExecError returns formatted error message", func(t *testing.T) {
		cmd := exec.Command("git", "diff")
		output := []byte("some output")
		err := fmt.Errorf("some error")

		expectedErrMsg := `command "git diff": some error: some output`
		actualErr := ExecError(cmd, output, err)
		require.Equal(t, expectedErrMsg, actualErr.Error(), "ExecError did not return the expected error message")
	})
}
