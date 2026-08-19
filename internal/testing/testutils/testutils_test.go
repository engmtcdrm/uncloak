package testutils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type zeroReader struct{}

func (z *zeroReader) Read(p []byte) (int, error) {
	return 0, nil
}

// Tests for [EmptyReader.Read] function.
func Test_EmptyReader_Read(t *testing.T) {
	t.Run("should return EOF for empty input", func(t *testing.T) {
		r := &EmptyReader{}
		buf := make([]byte, 10)
		n, err := r.Read(buf)
		assert.Equal(t, 0, n, "number of bytes read should be 0")
		assert.Equal(t, io.EOF, err, "error should be EOF")
	})
}

// Tests for [ErrorReader.Read] function.
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

// Tests for [CreatePTYWithSize] function.
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

// Tests for [ReadPTYOutput] function.
func Test_ReadPTYOutput(t *testing.T) {
	const expectedOutput = "test text"

	t.Run("should read output from the pseudo-terminal", func(t *testing.T) {
		master, slave := CreatePTYWithSize(t, 80, 30)

		_, _ = slave.Write([]byte(expectedOutput))
		_ = slave.Close()

		output := ReadPTYOutput(t, master, 1024)
		assert.Equal(t, expectedOutput, output)
	})

	t.Run("should read output from the pseudo-terminal if buffer size is less than or equal to 0", func(t *testing.T) {
		master, slave := CreatePTYWithSize(t, 80, 30)

		_, _ = slave.Write([]byte(expectedOutput))
		_ = slave.Close()

		output := ReadPTYOutput(t, master, 0)
		assert.Equal(t, expectedOutput, output)
	})

	t.Run("should stop reading when no bytes are read and no error is returned", func(t *testing.T) {
		output := ReadPTYOutput(t, &zeroReader{}, 1024)
		assert.Empty(t, output)
	})
}

// Tests for [SetStdout] function.
func Test_SetStdout(t *testing.T) {
	originalStdout := *os.Stdout

	tempFile := SetStdout(t)
	assert.Same(t, tempFile, os.Stdout, "os.Stdout should be set to the temp file")
	assert.NotEqual(t, originalStdout.Fd(), (*os.File)(os.Stdout).Fd(), "os.Stdout should be different from the original stdout")
}
