package gocover

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/mod/modfile"
)

// GoList represents the structure of the JSON output from `go list -m -json`.
type GoList struct {
	Path      string `json:"Path"`
	Main      bool   `json:"Main"`
	Dir       string `json:"Dir"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
	Module    string `json:"Module"`
}

// getGoList calls `go list -m -json` to retrieve module information and returns
// that information.
func getGoList() (*GoList, error) {
	cmd := exec.Command("go", "list", "-m", "-json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command: %q: %w: %s", strings.Join(cmd.Args, " "), err, string(output))
	}

	var mod *GoList
	err = json.Unmarshal(output, &mod)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(mod.GoMod)
	if err != nil {
		return nil, err
	}

	mod.Module = modfile.ModulePath(data)
	if mod.Module == "" {
		return nil, ErrNoModuleName
	}

	return mod, nil
}
