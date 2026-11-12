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
	// ErrEmptyCoverageFile indicates that the coverage file is empty.
	ErrEmptyCoverageFile = errors.New("empty coverage file")

	// ErrInvalidCoverageFile indicates that the coverage file has an invalid
	// format.
	ErrInvalidCoverageFile = errors.New("invalid coverage file format")

	// ErrInvalidCoverageLine indicates that a coverage line is invalid.
	ErrInvalidCoverageLine = errors.New("invalid coverage line")

	// ErrInvalidMode indicates that the coverage mode is invalid.
	ErrInvalidMode = errors.New("invalid coverage mode")

	// ErrMissingMode indicates that the coverage mode is missing.
	ErrMissingMode = errors.New("missing coverage mode")

	// ErrNoModuleName indicates that no module name was found in the go.mod
	// file.
	ErrNoModuleName = errors.New("no module name found in go.mod file")

	// ErrNoLinesToTest indicates that there are no lines to test in the
	// coverage file.
	ErrNoLinesToTest = errors.New("no lines to test")

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
