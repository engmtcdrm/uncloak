package testconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [CreateConfigFile] function.
func Test_CreateConfigFile(t *testing.T) {
	t.Run("should create a config file with the specified content", func(t *testing.T) {
		tempDir := t.TempDir()
		configFilePath := CreateConfigFile(t, tempDir, ValidYaml)

		content, err := os.ReadFile(configFilePath)
		require.NoError(t, err)
		require.Equal(t, string(content), string(ValidYaml))
	})
}
