package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

const validYaml = `
version: 0
exclusions:
  - "main.go"
  - "**/internal/**"
`

const invalidYaml = `
version: 0
unknown-field: true
`

// Tests for [Config.IsExclusionFile] method.
func Test_Config_IsExclusionFile(t *testing.T) {
	t.Run("should return true for exact match", func(t *testing.T) {
		cfg := &Config{
			Exclusions: []string{"main.go", "utils/*random.go"},
		}

		require.True(t, cfg.IsExclusionFile("main.go"))
		require.True(t, cfg.IsExclusionFile("utils/some_random.go"))
	})

	t.Run("should return true for glob match", func(t *testing.T) {
		cfg := &Config{
			Exclusions: []string{"**/internal/**"},
		}

		require.True(t, cfg.IsExclusionFile("src/internal/file.go"))
		require.True(t, cfg.IsExclusionFile("internal/file.go"))
	})

	t.Run("should return false for non-matching files", func(t *testing.T) {
		cfg := &Config{
			Exclusions: []string{"main.go", "**/internal/**"},
		}

		require.False(t, cfg.IsExclusionFile("utils/helper.go"))
		require.False(t, cfg.IsExclusionFile("src/external/file.go"))
	})
}

// Tests for [Load] function.
func Test_Load(t *testing.T) {
	t.Run("should load valid config file", func(t *testing.T) {
		tempDir := createTempConfigFile(t, validYaml)

		expectedConfig := &Config{
			Version:           0,
			CoverageThreshold: DefaultConfig.CoverageThreshold,
			Exclusions:        []string{"main.go", "**/internal/**"},
			GitDiffOptions:    DefaultConfig.GitDiffOptions, // Default value should be set
		}

		t.Chdir(tempDir)
		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("should return default config if no config file found", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, &DefaultConfig, cfg)
	})

	t.Run("should return error on unreadable config file", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping test on Windows due to permission issues with temp directories.")
		}

		tempDir := createTempConfigFile(t, validYaml)
		configFilePath := filepath.Join(tempDir, ".uncloak.yml")

		err := os.Chmod(configFilePath, 0000)
		require.NoError(t, err, "Failed to change file permissions")

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("should return error on invalid YAML", func(t *testing.T) {
		tempDir := createTempConfigFile(t, invalidYaml)

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("should return error if config validation fails", func(t *testing.T) {
		invalidConfigYaml := `
version: 0
coverage-threshold: -10.0
`
		tempDir := createTempConfigFile(t, invalidConfigYaml)

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
		require.Contains(t, err.Error(), "coverage-threshold must be between 0 and 100")
	})
}

// Tests for [validate] function.
func Test_validate(t *testing.T) {
	t.Run("should return no error for valid config", func(t *testing.T) {
		cfg := &DefaultConfig
		err := validate(cfg)
		require.NoError(t, err)
	})

	t.Run("should return error if coverage-threshold is negative", func(t *testing.T) {
		cfg := &Config{
			CoverageThreshold: -1.0,
		}
		err := validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "coverage-threshold must be between 0 and 100")
	})

	t.Run("should return error if coverage-threshold is greater than 100", func(t *testing.T) {
		cfg := &Config{
			CoverageThreshold: 101.0,
		}
		err := validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "coverage-threshold must be between 0 and 100")
	})

}

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, ".uncloak.yml")

	tmpFile, err := os.Create(configFilePath)
	require.NoError(t, err, "Failed to create temp file")

	_, err = tmpFile.Write([]byte(content))
	require.NoError(t, err, "Failed to write to temp file")
	require.NoError(t, tmpFile.Close())

	return tempDir
}
