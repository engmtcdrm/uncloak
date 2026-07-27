package testutils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

// Tests for [SetStdout] function.
func Test_SetStdout(t *testing.T) {
	originalStdout := *os.Stdout

	tempFile := SetStdout(t)
	assert.Same(t, tempFile, os.Stdout, "os.Stdout should be set to the temp file")
	assert.NotEqual(t, originalStdout.Fd(), (*os.File)(os.Stdout).Fd(), "os.Stdout should be different from the original stdout")
}
