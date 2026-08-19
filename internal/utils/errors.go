package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// ExecError formats an error message for a failed command execution, including
// the command, the error, and the output.
func ExecError(cmd *exec.Cmd, output []byte, err error) error {
	return fmt.Errorf("command %q: %w: %s", strings.Join(cmd.Args, " "), err, string(output))
}
