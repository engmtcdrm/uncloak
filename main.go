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
