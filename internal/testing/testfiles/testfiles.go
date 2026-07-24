package testfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateFile creates a file with the specified content at the given filePath.
func CreateFile(t *testing.T, filePath, content string) {
	t.Helper()

	tmpFile, err := os.Create(filePath)
	require.NoError(t, err, "Failed to create file")

	_, err = tmpFile.Write([]byte(content))
	require.NoError(t, err, "Failed to write to file")
	require.NoError(t, tmpFile.Close())
}

// CopyDir is a test helper function to copy all files from the
// source directory to destination directory. subdirectories are ignored.
func CopyDir(t *testing.T, srcDir string, destDir string) {
	t.Helper()

	entries, err := os.ReadDir(srcDir)
	require.NoError(t, err)

	err = os.MkdirAll(destDir, os.ModePerm)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() {
			CopyDir(t, filepath.Join(srcDir, entry.Name()), filepath.Join(destDir, entry.Name()))
			continue
		}

		CopyFile(t, filepath.Join(srcDir, entry.Name()), destDir)
	}
}

// CopyFile copies a single file from the specified source path to the specified
// destination directory.
func CopyFile(t *testing.T, srcPath string, destDir string) {
	t.Helper()

	srcFile, err := os.Open(srcPath)
	require.NoError(t, err)
	defer func() { _ = srcFile.Close() }()

	destPath := filepath.Join(destDir, filepath.Base(srcPath))
	destFile, err := os.Create(destPath)
	require.NoError(t, err)
	defer func() { _ = destFile.Close() }()

	_, err = destFile.ReadFrom(srcFile)
	require.NoError(t, err)
}
