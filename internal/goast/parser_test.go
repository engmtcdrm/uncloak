package goast

import (
	"context"
	"go/ast"
	"os"
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

	t.Run("should return nil for a non-Go source file", func(t *testing.T) {
		parsedFile, err := Parse("file.txt")
		require.NoError(t, err)
		require.Nil(t, parsedFile)
	})

	t.Run("should return nil for a directory", func(t *testing.T) {
		tempDir := t.TempDir()
		goDir := filepath.Join(tempDir, ".go")
		err := os.Mkdir(goDir, 0755)
		require.NoError(t, err)

		parsedFile, err := Parse(goDir)
		require.Error(t, err)
		require.Nil(t, parsedFile)
	})

	t.Run("should return an error if file is unreadable", func(t *testing.T) {
		t.Skip("Skipping test on Windows due to permission issues with temp directories.")

		tempDir := t.TempDir()
		file := filepath.Join(tempDir, "file.go")
		err := os.WriteFile(file, []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)

		// Make the file unreadable
		err = os.Chmod(file, 0000)
		require.NoError(t, err)

		parsedFile, err := Parse(file)
		require.Error(t, err)
		require.Nil(t, parsedFile)
	})

	t.Run("should return an error if file is not a go source file despite having .go extension", func(t *testing.T) {
		testfiles := t.TempDir()
		file := filepath.Join(testfiles, "file.go")
		err := os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)

		parsedFile, err := Parse(file)
		require.Error(t, err)
		require.Nil(t, parsedFile)
	})
}

// Tests for [parseFuncDecls] function.
func Test_parseFuncDecls(t *testing.T) {
}

// Tests for [receiverTypeName] function.
func Test_receiverTypeName(t *testing.T) {
	t.Run("should return empty string if fn.Recv is nil", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: nil,
		}
		require.Equal(t, "", receiverTypeName(fn))
	})

	t.Run("should return empty string if fn.Recv.List is empty", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: nil,
			},
		}
		require.Equal(t, "", receiverTypeName(fn))
	})

	t.Run("should return the receiver type name if fn.Recv.List has a valid receiver", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ident{Name: "MyType"},
					},
				},
			},
		}
		require.Equal(t, "MyType", receiverTypeName(fn))
	})

	t.Run("should return the receiver type name if fn.Recv.List has a pointer receiver", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: "MyType"},
						},
					},
				},
			},
		}
		require.Equal(t, "MyType", receiverTypeName(fn))
	})

	t.Run("should return empty string if fn.Recv.List has an unsupported receiver type", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.ArrayType{},
					},
				},
			},
		}
		require.Equal(t, "", receiverTypeName(fn))
	})
}
