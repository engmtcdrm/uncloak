package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testconfig"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/stretchr/testify/require"
)

// Tests for [Execute] function.
func Test_Execute(t *testing.T) {
	t.Run("Execute runs without error when in git repository", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		require.NoError(t, Execute())
	})

	t.Run("Execute runs with error if coverage is below default", func(t *testing.T) {
		tempDir, _ := testrepo.InitWithFileCopy(t)
		rmTestFile := filepath.Join(tempDir, "magic_100_test.go")
		err := os.Remove(rmTestFile)
		require.NoError(t, err)

		t.Chdir(tempDir)
		require.Error(t, Execute())
	})
}

// Tests for [cmd.run] method.
func Test_run(t *testing.T) {
	t.Run("should run the command without error when in git repository", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		c := &cmd{}
		localRootCmd := rootCmd

		err := c.run(localRootCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("should run without error when in git repository and with arguments", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(t)
		c := &cmd{}
		localRootCmd := rootCmd

		flags := map[string]string{
			"coverage-threshold": "1.0",
			"debug":              "true",
			"git-target-ref":     "origin/main",
			"verbose":            "true",
		}

		var err error
		c.verbose, err = strconv.ParseBool(flags["verbose"])
		require.NoError(t, err)

		// Set flags for the command
		for flag, value := range flags {
			err := localRootCmd.Flags().Set(flag, value)
			require.NoError(t, err)
		}

		err = c.run(localRootCmd, []string{})
		require.NoError(t, err)
	})

	t.Run("should error when config file is invalid", func(t *testing.T) {
		tempDir, _ := testrepo.Init(t)
		c := &cmd{}
		localRootCmd := rootCmd

		testconfig.CreateConfigFile(t, tempDir, testconfig.InvalidUnknownFieldYaml)

		err := c.run(localRootCmd, []string{})
		require.Error(t, err)
	})

	t.Run("should error when NewCodeCoverage error is not below threshold", func(t *testing.T) {
		c := &cmd{}
		localRootCmd := rootCmd

		t.Chdir(t.TempDir())
		err := c.run(localRootCmd, []string{})
		require.Error(t, err)
	})
}
