package testfiles

import (
	"os"
	"path/filepath"
	"testing"

	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/stretchr/testify/require"
)

// Tests for [AppendFile] function.
func Test_AppendFile(t *testing.T) {
	t.Run("should append content to an existing file", func(t *testing.T) {
		const initialContent = "Hello"
		const appendedContent = ", World!"
		const expectedContent = initialContent + appendedContent

		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "testfile.txt")
		CreateFile(t, filePath, initialContent)

		AppendFile(t, filePath, appendedContent)

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, expectedContent, string(content))
	})
}

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

		repoPath := testgit.GetTestRepoPath(t.Context(), t)

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

// Tests for [CopyFile] function.
func Test_CopyFile(t *testing.T) {
	t.Run("CopyFile copies a single file to the destination directory", func(t *testing.T) {
		tempDir := t.TempDir()
		srcFilePath := filepath.Join(tempDir, "source.txt")
		const content = "Hello, World!"
		CreateFile(t, srcFilePath, content)

		destDir := t.TempDir()
		CopyFile(t, srcFilePath, destDir)

		destFilePath := filepath.Join(destDir, "source.txt")
		readContent, err := os.ReadFile(destFilePath)
		require.NoError(t, err)
		require.Equal(t, content, string(readContent))
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
