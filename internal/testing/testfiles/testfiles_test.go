package testfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pp "github.com/engmtcdrm/go-prettyprint"
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
	ctx := context.Background()

	t.Run("CopyDir copies files from source to temp directory", func(t *testing.T) {
		tempDir := t.TempDir()

		repoPath := testgit.GetTestRepoPath(ctx, t)

		CopyDir(t, repoPath, tempDir)
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.NotEmpty(t, entries)
		require.Len(t, entries, 8)

		for _, entry := range entries {
			require.NotEqual(t, ".git", entry.Name())
		}
	})
}

// Tests for [ReadFile] function.
func Test_ReadFile(t *testing.T) {
	t.Run("should read the content of a file", func(t *testing.T) {
		const expectedContent = "Hello, World!"
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "testfile.txt")
		CreateFile(t, filePath, expectedContent)

		content := ReadFile(t, filePath)
		require.Equal(t, expectedContent, content)
	})
}

// Tests for [ReadFileWithANSIStrip] function.
func Test_ReadFileWithANSIStrip(t *testing.T) {
	t.Run("should read the content of a file and strip ANSI escape sequences", func(t *testing.T) {
		const expectedContent = "Hello, World!"
		var ansiContent = pp.Red(expectedContent)
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "testfile.txt")
		CreateFile(t, filePath, ansiContent)

		content := ReadFileWithANSIStrip(t, filePath)
		require.Equal(t, expectedContent, content)
	})

	t.Run("should handle files without ANSI escape sequences", func(t *testing.T) {
		const expectedContent = "Hello, World!"
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "testfile.txt")
		CreateFile(t, filePath, expectedContent)

		content := ReadFileWithANSIStrip(t, filePath)
		require.Equal(t, expectedContent, content)
	})
}
