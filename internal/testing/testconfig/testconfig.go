package testconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ValidYaml = `
version: 0
exclusions:
  - "main.go"
  - "**/internal/**"
`

	InvalidUnknownFieldYaml = `
version: 0
unknown-field: true
`

	InvalidEmptyYaml = ""

	InvalidCoverageThresholdYaml = `
version: 0
coverage-threshold: -10.0
`
)

func CreateTempConfigFile(t *testing.T, tempDir, content string) string {
	t.Helper()

	if tempDir == "" {
		tempDir = t.TempDir()
	}
	configFilePath := filepath.Join(tempDir, ".uncloak.yml")

	tmpFile, err := os.Create(configFilePath)
	require.NoError(t, err, "Failed to create temp file")

	_, err = tmpFile.Write([]byte(content))
	require.NoError(t, err, "Failed to write to temp file")
	require.NoError(t, tmpFile.Close())

	return tempDir
}
