//go:build linux
// +build linux

package utils

import (
	"os/exec"
	"strings"
	"syscall"
)

func WasExecTerminated(err error) bool {
	switch {
	case err == nil:
		return false
	case strings.Contains(err.Error(), "context canceled"):
		return true
	case strings.Contains(err.Error(), "killed"):
		return true
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.Signaled() && status.Signal() == syscall.SIGKILL
		}
	}

	return false
}
