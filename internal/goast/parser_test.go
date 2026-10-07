package goast

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
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
		if runtime.GOOS == "windows" {
			t.Skip("Skipping test on Windows due to permission issues with temp directories.")
		}

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
	t.Run("should return empty funcDecls and funcOrder if astFile has empty Decls", func(t *testing.T) {
		astFile := &ast.File{
			Decls: nil,
		}
		fileSet := token.NewFileSet()
		funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)
		require.Empty(t, funcDecls)
		require.Empty(t, funcOrder)
	})

	t.Run("should return funcDecls and funcOrder for a file with one function", func(t *testing.T) {
		src := `
package main

func Foo() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)
		require.Len(t, funcDecls, 1)
		require.Contains(t, funcDecls, "Foo")
		require.Equal(t, []string{"Foo"}, funcOrder)
	})

	t.Run("should return funcDecls and funcOrder with a receiver prepended to function name", func(t *testing.T) {
		src := `
package main

type MyType struct{}

func (m MyType) Foo() {}

type MyType2 struct {}

func (m *MyType2) Foo() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)
		require.Len(t, funcDecls, 2)
		require.Contains(t, funcDecls, "MyType.Foo")
		require.Contains(t, funcDecls, "MyType2.Foo")
		require.Equal(t, []string{"MyType.Foo", "MyType2.Foo"}, funcOrder)
	})

	t.Run("should use physical line numbers when the file has a line directive", func(t *testing.T) {
		src := `package main
//line mapped.go:100
func Foo() {
}
`
		expectedFuncDecl := &FuncDecl{
			Name:  "Foo",
			Lines: []int{3, 4},
			Start: 3,
			End:   4,
			Body: FuncBody{
				Lines: []int{3, 4},
				Start: 3,
				End:   4,
			},
		}

		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)
		require.NotEmpty(t, astFile.Decls)
		require.Equal(t, 100, fileSet.Position(astFile.Decls[0].Pos()).Line)

		funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)
		require.Equal(t, []string{"Foo"}, funcOrder)
		require.Equal(t, expectedFuncDecl, funcDecls["Foo"])
	})

	t.Run("should handle init functions with unique names", func(t *testing.T) {
		src := `
package main

func init() {}
func init() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)
		require.Len(t, funcDecls, 2)
		require.Equal(t, 2, len(funcOrder))
		require.Contains(t, funcOrder[0], "init@")
		require.Contains(t, funcOrder[1], "init@")
	})
}

// Tests for [receiverTypeName] function.
func Test_receiverTypeName(t *testing.T) {
	t.Run("should return empty string if fn.Recv is nil", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: nil,
		}
		require.Empty(t, receiverTypeName(fn))
	})

	t.Run("should return empty string if fn.Recv.List is empty", func(t *testing.T) {
		fn := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: nil,
			},
		}
		results := receiverTypeName(fn)
		require.Empty(t, results)
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
		require.Empty(t, receiverTypeName(fn))
	})
}

// Tests for [receiverBaseTypeName] function.
func Test_receiverBaseTypeName(t *testing.T) {
	t.Run("should return the name for an ident expression", func(t *testing.T) {
		expr := &ast.Ident{Name: "MyType"}
		require.Equal(t, "MyType", receiverBaseTypeName(expr))
	})

	t.Run("should return the name for a pointer receiver with an ident expression", func(t *testing.T) {
		expr := &ast.StarExpr{X: &ast.Ident{Name: "MyType"}}
		require.Equal(t, "MyType", receiverBaseTypeName(expr))
	})

	t.Run("should recurse for a pointer receiver with a non-ident expression", func(t *testing.T) {
		expr := &ast.StarExpr{
			X: &ast.IndexExpr{
				X: &ast.Ident{Name: "MyType"},
			},
		}
		require.Equal(t, "MyType", receiverBaseTypeName(expr))
	})

	t.Run("should recurse for an indexed (generic) type expression", func(t *testing.T) {
		expr := &ast.IndexExpr{
			X: &ast.Ident{Name: "MyType"},
		}
		require.Equal(t, "MyType", receiverBaseTypeName(expr))
	})

	t.Run("should recurse for a generic type with multiple type parameters", func(t *testing.T) {
		expr := &ast.IndexListExpr{
			X: &ast.Ident{Name: "MyType"},
		}
		require.Equal(t, "MyType", receiverBaseTypeName(expr))
	})

	t.Run("should return empty string for an unsupported expression type", func(t *testing.T) {
		expr := &ast.ArrayType{}
		require.Empty(t, receiverBaseTypeName(expr))
	})
}
