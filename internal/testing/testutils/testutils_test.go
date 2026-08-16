package testutils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [EmptyReader.Read] method.
func Test_EmptyReader_Read(t *testing.T) {
	t.Run("should return EOF for empty input", func(t *testing.T) {
		r := &EmptyReader{}
		buf := make([]byte, 10)
		n, err := r.Read(buf)
		assert.Equal(t, 0, n, "number of bytes read should be 0")
		assert.Equal(t, io.EOF, err, "error should be EOF")
	})
}

// Tests for [ErrorReader.Read] method.
func Test_ErrorReader_Read(t *testing.T) {
	t.Run("should return unexpected EOF error", func(t *testing.T) {
		r := &ErrorReader{}
		buf := make([]byte, 10)
		n, err := r.Read(buf)
		assert.Equal(t, 0, n, "number of bytes read should be 0")
		assert.Equal(t, io.ErrUnexpectedEOF, err, "error should be ErrUnexpectedEOF")
	})
}

// Tests for [CreatePTY] function.
func Test_CreatePTY(t *testing.T) {
	t.Run("should create a pseudo-terminal pair", func(t *testing.T) {
		master, slave := CreatePTY(t)
		assert.NotNil(t, master, "master should not be nil")
		assert.NotNil(t, slave, "slave should not be nil")
	})
}

// Tets for [CreatePTYWithSize] function.
func Test_CreatePTYWithSize(t *testing.T) {
	t.Run("should create a pseudo-terminal pair with specified size", func(t *testing.T) {
		columns, rows := 80, 24
		master, slave := CreatePTYWithSize(t, columns, rows)
		t.Cleanup(func() {
			require.NoError(t, master.Close(), "Failed to close master pty")
			require.NoError(t, slave.Close(), "Failed to close slave pty")
		})
		assert.NotNil(t, master, "master should not be nil")
		assert.NotNil(t, slave, "slave should not be nil")
	})
}

// Tests for [SetStdout] function.
func Test_SetStdout(t *testing.T) {
	originalStdout := *os.Stdout

	tempFile := SetStdout(t)
	assert.Same(t, tempFile, os.Stdout, "os.Stdout should be set to the temp file")
	assert.NotEqual(t, originalStdout.Fd(), (*os.File)(os.Stdout).Fd(), "os.Stdout should be different from the original stdout")
}
