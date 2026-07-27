package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func ExecError(cmd *exec.Cmd, output []byte, err error) error {
	return fmt.Errorf("command %q: %w: %s", strings.Join(cmd.Args, " "), err, string(output))
}
