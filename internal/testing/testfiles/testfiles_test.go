package testfiles

import (
	"os"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/stretchr/testify/require"
)

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
