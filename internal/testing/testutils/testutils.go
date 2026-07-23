package testutils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// EmptyReader is a mock reader that simulates an empty input.
type EmptyReader struct{}

func (e *EmptyReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// ErrorReader is a mock reader that simulates an error when reading.
type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

// SetStdout is a helper function that stores the originl [os.Stdout], replaces
// it with a temporary file, and restores the original [os.Stdout] after the
// test finishes. It returns the temporary file.
func SetStdout(t *testing.T) *os.File {
	t.Helper()

	originalStdout := *os.Stdout
	t.Cleanup(func() {
		os.Stdout = &originalStdout
	})

	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "stdout")
	require.NoError(t, err, "Failed to create temp file: %v", err)
	t.Cleanup(func() {
		require.NoError(t, tempFile.Close(), "Failed to close temp file: %v", err)
	})

	os.Stdout = tempFile

	return tempFile
}
