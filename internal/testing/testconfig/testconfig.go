package testconfig

import (
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
)

type TestYamlFile string

const (
	ValidYaml TestYamlFile = `
version: 0
exclusions:
  - "main.go"
  - "**/internal/**"
`

	InvalidEmptyYaml TestYamlFile = ""

	InvalidCoverageThresholdYaml TestYamlFile = `
version: 0
coverage-threshold: -10.0
`

	InvalidUnknownFieldYaml TestYamlFile = `
version: 0
unknown-field: true
`
)

// CreateConfigFile creates a .uncloak.yml configuration file with the specified
// content in the given directory.
func CreateConfigFile(t *testing.T, dir string, content TestYamlFile) string {
	t.Helper()

	configFilePath := filepath.Join(dir, ".uncloak.yml")

	testfiles.CreateFile(t, configFilePath, string(content))

	return configFilePath
}
