package testrepo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [Init] function.
func Test_Init(t *testing.T) {
	tempDir, stdoutFile := Init(t)
	require.NotEmpty(t, tempDir)
	require.NotEmpty(t, stdoutFile)
}

// Tests for [InitWithFileCopy] function.
func Test_InitWithFileCopy(t *testing.T) {
	tempDir, stdoutFile := InitWithFileCopy(t)
	require.NotEmpty(t, tempDir)
	require.NotEmpty(t, stdoutFile)
}
