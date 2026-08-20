package config

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/engmtcdrm/uncloak/internal/gocover"
	"github.com/engmtcdrm/uncloak/internal/testing/testconfig"
	"github.com/stretchr/testify/require"
)

// Tests for [Config.IsExclusionFile] function.
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
		tempDir := t.TempDir()
		_ = testconfig.CreateConfigFile(t, tempDir, testconfig.ValidYaml)

		expectedConfig := &Config{
			Version:           0,
			CoverageThreshold: DefaultConfig.CoverageThreshold,
			Exclusions:        []string{"main.go", "**/internal/**"},
			GitDiffOptions:    DefaultConfig.GitDiffOptions, // Default value should be set
			GoTestOptions:     DefaultConfig.GoTestOptions,  // Default value should be set
		}

		t.Chdir(tempDir)
		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("should load go-test options from yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = testconfig.CreateConfigFile(t, tempDir, testconfig.ValidGoTestOptionsYaml)

		expectedConfig := &Config{
			Version:           0,
			CoverageThreshold: DefaultConfig.CoverageThreshold,
			GoTestOptions: gocover.Options{
				Count:   3,
				Timeout: 30 * time.Second,
				Verbose: true,
			},
			GitDiffOptions: DefaultConfig.GitDiffOptions,
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

		tempDir := t.TempDir()
		configFilePath := testconfig.CreateConfigFile(t, tempDir, testconfig.ValidYaml)

		err := os.Chmod(configFilePath, 0000)
		require.NoError(t, err, "Failed to change file permissions")

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("should return error on invalid YAML", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = testconfig.CreateConfigFile(t, tempDir, testconfig.InvalidUnknownFieldYaml)

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("should return error if config validation fails", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = testconfig.CreateConfigFile(t, tempDir, testconfig.InvalidCoverageThresholdYaml)

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
		require.Contains(t, err.Error(), "coverage-threshold must be between 0 and 100")
	})

	t.Run("should return error if config file is empty", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = testconfig.CreateConfigFile(t, tempDir, testconfig.InvalidEmptyYaml)

		t.Chdir(tempDir)
		cfg, err := Load()
		require.Error(t, err)
		require.Nil(t, cfg)
		require.ErrorIs(t, err, ErrConfigFileEmpty)
	})
}

// Tests for [validate] function.
func Test_validate(t *testing.T) {
	t.Run("should return no error for valid config", func(t *testing.T) {
		cfg := &DefaultConfig
		err := Validate(cfg)
		require.NoError(t, err)
	})

	t.Run("should return error if version is not 0", func(t *testing.T) {
		cfg := &Config{
			Version: 1,
		}
		err := Validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported config version")
	})

	t.Run("should return error if coverage-threshold is negative", func(t *testing.T) {
		cfg := &Config{
			CoverageThreshold: -1.0,
		}
		err := Validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "coverage-threshold must be between 0 and 100")
	})

	t.Run("should return error if coverage-threshold is greater than 100", func(t *testing.T) {
		cfg := &Config{
			CoverageThreshold: 101.0,
		}
		err := Validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "coverage-threshold must be between 0 and 100")
	})

	t.Run("should return error if go-test.timeout is negative", func(t *testing.T) {
		cfg := &Config{
			GoTestOptions: gocover.Options{
				Timeout: -1,
			},
		}
		err := Validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "go-test.timeout must be non-negative")
	})

	t.Run("should return no error if go-test.timeout is zero", func(t *testing.T) {
		cfg := &Config{
			GoTestOptions: gocover.Options{
				Timeout: 0,
			},
		}
		err := Validate(cfg)
		require.NoError(t, err)
	})

	t.Run("should return no error if go-test.timeout is positive", func(t *testing.T) {
		cfg := &Config{
			GoTestOptions: gocover.Options{
				Timeout: 10,
			},
		}
		err := Validate(cfg)
		require.NoError(t, err)
	})

	t.Run("should return error if go-test.count is negative", func(t *testing.T) {
		cfg := &Config{
			GoTestOptions: gocover.Options{
				Count: -1,
			},
		}
		err := Validate(cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "go-test.count must be non-negative")
	})

	t.Run("should return no error if go-test.count is zero", func(t *testing.T) {
		cfg := &Config{
			GoTestOptions: gocover.Options{
				Count: 0,
			},
		}
		err := Validate(cfg)
		require.NoError(t, err)
	})

	t.Run("should return no error if go-test.count is positive", func(t *testing.T) {
		cfg := &Config{
			GoTestOptions: gocover.Options{
				Count: 5,
			},
		}
		err := Validate(cfg)
		require.NoError(t, err)
	})
}
