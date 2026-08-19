// Package app provides application information such as name, version,
// description, etc.
package app

import (
	"runtime/debug"
)

const (
	// Name of the application.
	Name = "uncloak"

	// LongDesc provides a detailed description of the application.
	LongDesc = "A CLI tool for analyzing new code coverage in Go files based on git diffs and coverage profiles."

	// ShortDesc provides a brief description of the application.
	ShortDesc = "A CLI tool for analyzing new code coverage in Go files"

	// RepoURL is the URL of the application's repository.
	RepoURL = "https://github.com/engmtcdrm/uncloak"
)

var (
	// Version of the application.
	Version = ""

	readBuildInfo = debug.ReadBuildInfo
)

func init() {
	setVersion(&Version)
}

// setVersion sets the Version.
func setVersion(version *string) {
	if *version != "" {
		return
	}

	info, ok := readBuildInfo()
	if !ok {
		return
	}

	*version = info.Main.Version
}
