package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/app"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [Execute] function.
func Test_Execute(t *testing.T) {
	// Simple test helper to initialize a new root command and set required
	// flag(s). We want this due to the nature of the rootCmd being global
	// within this package. Otherwise flags could carry over.
	initValidCmd := func(t *testing.T) {
		t.Helper()

		rootCmd = newRootCmd()

		err := rootCmd.Flags().Set("target-ref", gitdiff.LocalMain)
		require.NoError(t, err)
	}

	t.Run("should run without error when in git repository", func(t *testing.T) {
		initValidCmd(t)
		_, _ = testrepo.InitWithFileCopy(t.Context(), t)

		err := Execute()
		require.NoError(t, err)
	})

	t.Run("should return error if coverage is below default", func(t *testing.T) {
		_, _ = testutils.SetOutputStreams(t)

		initValidCmd(t)
		tempDir, _ := testrepo.InitWithFileCopy(t.Context(), t)

		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		t.Chdir(tempDir)
		err = Execute()
		require.Error(t, err)

		coverageThresholdError := &coverageThresholdError{}
		require.ErrorAs(t, err, &coverageThresholdError)
	})

	t.Run("should return error if target-ref is not set", func(t *testing.T) {
		_, _ = testutils.SetOutputStreams(t)

		rootCmd = newRootCmd()
		err := Execute()
		require.Error(t, err)
		require.ErrorIs(t, err, errGitTargetRef)
	})
}

// Tests for [init] function.
func Test_init(t *testing.T) {
	t.Run("should set rootCmd", func(t *testing.T) {
		require.NotNil(t, rootCmd)
	})
}

// Tests for [newRootCmd] function.
func Test_newRootCmd(t *testing.T) {
	t.Run("should return a non-nil root command", func(t *testing.T) {
		cmd := newRootCmd()
		require.NotNil(t, cmd)
		assert.Equal(t, app.Name, cmd.Use)
		assert.Equal(t, app.ShortDesc, cmd.Short)
		assert.Equal(t, app.LongDesc, cmd.Long)
		assert.Equal(t, app.Name+" -t main", cmd.Example)
		assert.Equal(t, app.Version, cmd.Version)
	})
}
