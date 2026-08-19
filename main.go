// Package main provides the entry point for the uncloak command-line
// application.
package main

import (
	"os"

	"github.com/engmtcdrm/uncloak/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
