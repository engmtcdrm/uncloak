package testrepo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [Init] function.
func Test_Init(t *testing.T) {
	ctx := context.Background()
	tempDir, stdoutFile := Init(ctx, t)
	require.NotEmpty(t, tempDir)
	require.NotEmpty(t, stdoutFile)
}

// Tests for [InitWithFileCopy] function.
func Test_InitWithFileCopy(t *testing.T) {
	ctx := context.Background()
	tempDir, stdoutFile := InitWithFileCopy(ctx, t)
	require.NotEmpty(t, tempDir)
	require.NotEmpty(t, stdoutFile)
}
