package gocover

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/engmtcdrm/uncloak/internal/utils"
)

var (
	ErrNoModuleName         = errors.New("no module name found in go.mod file")
	ErroInvalidCoverageFile = errors.New("invalid coverage file format")

	panicText = []byte("panic:")
)

func handleParseError(cmd *exec.Cmd, output []byte, err error) error {
	if execErr, ok := err.(*exec.ExitError); ok {
		output = append(output, execErr.Stderr...)
		if bytes.Contains(output, panicText) {
			return fmt.Errorf("command %q: unhandled panic in test: %s",
				strings.Join(cmd.Args, " "),
				string(output),
			)
		}

		return utils.ExecError(cmd, output, err)
	}

	return utils.ExecError(cmd, output, err)
}
