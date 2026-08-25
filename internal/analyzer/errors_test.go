package analyzer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [newTaskCanceledError] function.
func Test_newTaskCanceledError(t *testing.T) {
	t.Run("should create a new taskCanceledError with the given error", func(t *testing.T) {
		underlyingErr := errors.New("original error")
		err := newTaskCanceledError(underlyingErr)
		require.Error(t, err)
		require.Error(t, err.e)
		assert.Equal(t, underlyingErr, err.e)
	})

	t.Run("should create a new taskCanceledError with nil if the given error is nil", func(t *testing.T) {
		err := newTaskCanceledError(nil)
		require.Error(t, err)
		assert.NoError(t, err.e)
	})
}

// Tests for [taskCanceledError.Error] function.
func Test_taskCanceledError_Error(t *testing.T) {
	t.Run("should return the error message of the underlying error", func(t *testing.T) {
		const expectedMessage = "original error"
		err := errors.New(expectedMessage)
		taskErr := newTaskCanceledError(err)
		require.Error(t, taskErr)
		assert.Equal(t, taskErr.e, err)
		assert.Equal(t, expectedMessage, taskErr.Error())
	})

	t.Run("should return an empty string if the underlying error is nil", func(t *testing.T) {
		taskErr := newTaskCanceledError(nil)
		require.Error(t, taskErr)
		require.NoError(t, taskErr.e)
	})
}

// Tests for [taskCanceledError.Unwrap] function.
func Test_taskCanceledError_Unwrap(t *testing.T) {
	t.Run("should return the underlying error", func(t *testing.T) {
		underlyingErr := errors.New("original error")
		taskErr := newTaskCanceledError(underlyingErr)
		require.Error(t, taskErr)
		assert.Equal(t, underlyingErr, taskErr.Unwrap())
	})

	t.Run("should return nil if the underlying error is nil", func(t *testing.T) {
		taskErr := newTaskCanceledError(nil)
		require.Error(t, taskErr)
		assert.NoError(t, taskErr.Unwrap())
	})
}
