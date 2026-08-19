package gocover

// Options represents the configuration options for Go coverage analysis.
type Options struct {
	Rerun bool `yaml:"rerun"` // Whether to rerun tests to generate coverage data.
}
