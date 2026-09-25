package goast

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/stretchr/testify/require"
)

// Tests for [Parse] function.
func Test_Parse(t *testing.T) {
	t.Run("should parse a valid Go source file", func(t *testing.T) {
		ctx := context.Background()
		rootDir := testgit.RootDir(ctx, t)
		file := filepath.Join(rootDir, testgit.TestRepoDir, "magic.go")
		parsedFile, err := Parse(file)
		require.NoError(t, err)
		require.NotNil(t, parsedFile)
	})

	t.Run("should return an error for an invalid Go source file", func(t *testing.T) {
		parsedFile, err := Parse("invalid.go")
		require.Error(t, err)
		require.Nil(t, parsedFile)
	})
}

// Tests for [parseFuncDecls] function.
func Test_parseFuncDecls(t *testing.T) {
}

// Tests for [receiverTypeName] function.
func Test_receiverTypeName(t *testing.T) {
}
