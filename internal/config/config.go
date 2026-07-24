package config

import (
	"bytes"
	"os"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-yaml"

	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
)

var (
	DefaultConfig = Config{
		Version:           0,
		CoverageThreshold: 80.0,
		GitDiffOptions:    gitdiff.DefaultOptions,
		GoTestOptions:     gocover.Options{},
	}
	configFilenames = []string{".uncloak.yml", ".uncloak.yaml"}
)

type Config struct {
	Version           int             `yaml:"version"`            // Version of the config file format.
	Exclusions        []string        `yaml:"exclusions"`         // List of file patterns to exclude from analysis.
	CoverageThreshold float64         `yaml:"coverage-threshold"` // Minimum coverage threshold.
	GitDiffOptions    gitdiff.Options `yaml:"git"`                // Git-related configuration.
	GoTestOptions     gocover.Options `yaml:"test"`               // Go Test-related configuration.

	Debug bool // Not configurable via YAML, used for enabling debug output.
}

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
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

	return &DefaultConfig, nil
}

func load(r *bytes.Reader) (*Config, error) {
	cfg := DefaultConfig
	dec := yaml.NewDecoder(r, yaml.DisallowUnknownField())
	err := dec.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.CoverageThreshold < 0.0 || cfg.CoverageThreshold > 100.0 {
		return &ConfigError{Message: "coverage-threshold must be between 0 and 100"}
	}

	return nil
}
