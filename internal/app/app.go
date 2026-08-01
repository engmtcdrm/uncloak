// Application information such as name, version, description, etc.
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

	// RepoUrl is the URL of the application's repository.
	RepoUrl = "https://github.com/engmtcdrm/uncloak"
)

var (
	// Version of the application.
	Version = "dev"
)

func init() {
	// Only override the Version variable if it is still set to "dev"
	if Version == "dev" {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}

		Version = info.Main.Version
	}
}
