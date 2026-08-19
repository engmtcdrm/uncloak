package testutils

import (
	"bytes"
	"errors"
	"io"
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/creack/pty"
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

// CreatePTY creates a pseudo-terminal pair for testing terminal interactions.
// The returned master and slave *os.File can be used to simulate terminal input
// and output in tests. The master end can be used to write input as if typed by
// a user, while the slave end can be used to read output from the terminal.
// Both files are automatically closed after the test completes.
//
// Lovely stolen from https://github.com/engmtcdrm/go-pardon testutils package.
func CreatePTY(t *testing.T) (master *os.File, slave *os.File) {
	t.Helper()

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("pty is not supported on Windows or macOS")
	}

	m, s, err := pty.Open()
	require.NoError(t, err, "failed to open pty")

	t.Cleanup(func() {
		_ = m.Close()
		_ = s.Close()
	})

	return m, s
}

// CreatePTYWithSize creates a pseudo-terminal pair with the specified size for
// testing terminal interactions. The returned master and slave *os.File can be
// used to simulate terminal input and output in tests that require specific
// terminal dimensions. Both files are automatically closed after the test
// completes.
//
// Lovely stolen from https://github.com/engmtcdrm/go-pardon testutils package.
func CreatePTYWithSize(t *testing.T, columns, rows int) (master *os.File, slave *os.File) {
	t.Helper()

	master, slave = CreatePTY(t)

	err := pty.Setsize(slave, &pty.Winsize{Cols: uint16(columns), Rows: uint16(rows)})
	require.NoError(t, err, "failed to set pty size")

	return master, slave
}

// ReadPTYOutput reads all available output from the provided reader until EOF
// is reached.
func ReadPTYOutput(t *testing.T, ptyFile io.Reader, bufferSize int) string {
	t.Helper()

	if bufferSize <= 0 {
		bufferSize = 1024
	}

	var output bytes.Buffer
	buffer := make([]byte, bufferSize)
	for {
		n, readErr := ptyFile.Read(buffer)
		if n > 0 {
			_, _ = output.Write(buffer[:n])
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) || errors.Is(readErr, syscall.EIO) {
				break
			}
			require.NoError(t, readErr, "reading output should not return an error")
		}

		if n == 0 {
			break
		}
	}

	return output.String()
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
