package testfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/stretchr/testify/require"
)

// Tests for [CreateFile] function.
func Test_CreateFile(t *testing.T) {
	t.Run("should create a file with the specified content", func(t *testing.T) {
		const expectedContent = "Hello, World!"
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "testfile.txt")
		CreateFile(t, filePath, expectedContent)

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, expectedContent, string(content))
	})
}

// Tests for [CopyDir] function.
func Test_CopyDir(t *testing.T) {
	t.Run("CopyDir copies files from source to temp directory", func(t *testing.T) {
		tempDir := t.TempDir()

		repoPath := testgit.GetTestRepoPath(t)

		CopyDir(t, repoPath, tempDir)
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.NotEmpty(t, entries)
		require.Len(t, entries, 8)
	})
}
