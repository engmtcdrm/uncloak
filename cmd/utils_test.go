package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for [getSemVer] function.
func Test_getSemVer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid semver with v prefix",
			input:    "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "valid semver without v prefix",
			input:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "valid semver with leading zeros",
			input:    "v01.02.03",
			expected: "01.02.03",
		},
		{
			name:     "valid semver with large numbers",
			input:    "v123.456.789",
			expected: "123.456.789",
		},
		{
			name:     "invalid semver - too many parts",
			input:    "v1.2.3.4",
			expected: "v1.2.3.4",
		},
		{
			name:     "invalid semver - too few parts",
			input:    "v1.2",
			expected: "v1.2",
		},
		{
			name:     "invalid semver - non-numeric",
			input:    "v1.2.a",
			expected: "v1.2.a",
		},
		{
			name:     "invalid semver - with text",
			input:    "v1.2.3-beta",
			expected: "v1.2.3-beta",
		},
		{
			name:     "non-semver string",
			input:    "main",
			expected: "main",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "just v",
			input:    "v",
			expected: "v",
		},
		{
			name:     "double v prefix",
			input:    "vv1.2.3",
			expected: "vv1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSemVer(tt.input)
			assert.Equalf(t, tt.expected, result, "getSemVer(%q) = %q, expected %q", tt.input, result, tt.expected)
		})
	}
}
