package header

import (
	"os"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/require"
)

// Tests for [PrintHeader] function.
func Test_PrintHeader(t *testing.T) {
	t.Run("should print header to stdout", func(t *testing.T) {
		tempFile := testutils.SetStdout(t)

		PrintHeader()

		content, err := os.ReadFile(tempFile.Name())
		require.NoError(t, err)
		require.Greater(t, len(content), 0, "Expected header output, got empty string")
	})
}
