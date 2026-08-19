package testconfig

import (
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
)

// TestYamlFile represents a test YAML configuration file.
type TestYamlFile string

const (
	// ValidYaml represents a valid test YAML configuration file.
	ValidYaml TestYamlFile = `
version: 0
exclusions:
  - "main.go"
  - "**/internal/**"
`

	// InvalidEmptyYaml represents an invalid test YAML configuration file that
	// is empty.
	InvalidEmptyYaml TestYamlFile = ""

	// InvalidCoverageThresholdYaml represents an invalid test YAML
	// configuration file with a negative coverage threshold.
	InvalidCoverageThresholdYaml TestYamlFile = `
version: 0
coverage-threshold: -10.0
`

	// InvalidUnknownFieldYaml represents an invalid test YAML configuration
	// file with an unknown field.
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
