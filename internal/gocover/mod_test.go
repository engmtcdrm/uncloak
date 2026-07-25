package gocover

import (
	"os"
	"runtime"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testfiles"
	"github.com/stretchr/testify/require"
)

// Tests for [getGoList] function.
func Test_getGoList(t *testing.T) {
	t.Run("valid module name return", func(t *testing.T) {
		moduleName, err := getGoList()
		require.NoError(t, err)
		require.NotEmpty(t, moduleName, "Expected module name, got empty string")
	})

	t.Run("error when go.mod is missing", func(t *testing.T) {
		t.Chdir(t.TempDir())
		moduleName, err := getGoList()
		require.Error(t, err)
		require.Empty(t, moduleName)
	})

	t.Run("error when os.Getwd fails", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping test on Windows due to os.Getwd behavior")
		}

		// Create a temporary directory, change to it, then delete it to
		// simulate os.Getwd failure
		tempDir := t.TempDir()
		t.Chdir(tempDir)
		err := os.RemoveAll(tempDir)
		require.NoError(t, err, "Failed to remove temporary directory")

		moduleName, err := getGoList()
		require.Error(t, err)
		require.Empty(t, moduleName)
	})

	t.Run("error when module path is messed up", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		testfiles.CreateFile(t, "go.mod", "module\n\ngo 1.26")

		moduleName, err := getGoList()
		require.Error(t, err)
		require.Empty(t, moduleName)
	})
}
