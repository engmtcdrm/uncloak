// Package config provides functionality for loading and validating the
// configuration of the uncloak application. It supports reading configuration
// from YAML files, applying default values, and validating the configuration
// fields.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-yaml"

	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
)

var (
	// DefaultConfig is the default configuration used when no config file is
	// found.
	DefaultConfig = Config{
		Version:           0,
		CoverageThreshold: 80.0,
		GitDiffOptions:    gitdiff.DefaultOptions,
		GoTestOptions:     gocover.DefaultOptions,
	}

	// ErrConfigFileEmpty is returned when the configuration file is empty.
	ErrConfigFileEmpty = errors.New("config file is empty. Delete it or add valid configuration content")

	configFilenames = []string{".uncloak.yml", ".uncloak.yaml"}
)

// Config represents the configuration of the uncloak application.
type Config struct {
	Version           int             `yaml:"version"`            // Version of the config file format.
	Exclusions        []string        `yaml:"exclusions"`         // List of file patterns to exclude from analysis.
	CoverageThreshold float64         `yaml:"coverage-threshold"` // Minimum coverage threshold.
	GoTestOptions     gocover.Options `yaml:"go-test"`            // Go test options for coverage analysis.

	Debug          bool            // Not configurable via YAML, used for enabling debug output.
	GitDiffOptions gitdiff.Options // Git-related configuration.
}

// ValidationError represents an error related to the configuration validation.
type ValidationError struct {
	Message string
}

// Error implements the error interface for [ValidationError].
func (e *ValidationError) Error() string {
	return "config validation: " + e.Message
}

// IsExclusionFile checks if the given file matches any of the exclusion
// patterns in the config.
func (cfg *Config) IsExclusionFile(file string) bool {
	for _, exclusion := range cfg.Exclusions {
		if file == exclusion {
			return true
		}

		matched, _ := doublestar.Match(exclusion, file)
		if matched {
			return true
		}
	}

	return false
}

// Load reads the configuration from the first available config file and returns
// a Config struct.
func Load() (*Config, error) {
	for _, name := range configFilenames {
		data, err := os.ReadFile(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		return load(bytes.NewReader(data))
	}

	cfg := DefaultConfig
	return &cfg, nil
}

// load reads the configuration from the provided reader and returns a [Config]
// struct. If the configuration is invalid, it returns an error.
func load(r *bytes.Reader) (*Config, error) {
	cfg := DefaultConfig
	dec := yaml.NewDecoder(r, yaml.DisallowUnknownField())
	err := dec.Decode(&cfg)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrConfigFileEmpty
		}

		return nil, err
	}

	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks the configuration for any invalid values and returns an error
// if any are found.
func Validate(cfg *Config) error {
	if cfg.GoTestOptions.Count < 0 {
		message := fmt.Sprintf("go-test.count must be non-negative, got %d", cfg.GoTestOptions.Count)
		return &ValidationError{Message: message}
	}

	if cfg.CoverageThreshold < 0.0 || cfg.CoverageThreshold > 100.0 {
		message := fmt.Sprintf("coverage-threshold must be between 0 and 100, got %f", cfg.CoverageThreshold)
		return &ValidationError{Message: message}
	}

	if cfg.GoTestOptions.Timeout < 0 {
		message := fmt.Sprintf("go-test.timeout must be non-negative, got %v", cfg.GoTestOptions.Timeout)
		return &ValidationError{Message: message}
	}

	return nil
}
