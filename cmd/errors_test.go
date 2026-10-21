package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [newCoverageThresholdError] function.
func Test_newCoverageThresholdError(t *testing.T) {
	t.Run("should return a coverage threshold error", func(t *testing.T) {
		expected := &coverageThresholdError{}

		actual := newCoverageThresholdError(33.3, 80.0)
		require.Error(t, actual)
		require.ErrorAs(t, actual, &expected)
	})
}

// Tests for [coverageThresholdError.Error] function.
func Test_coverageThresholdError_Error(t *testing.T) {
	t.Run("should return the correct error message", func(t *testing.T) {
		err := newCoverageThresholdError(33.3, 80.0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "33.3")
		require.Contains(t, err.Error(), "80.0")
	})
}
