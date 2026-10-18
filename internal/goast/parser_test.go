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
	"github.com/stretchr/testify/assert"
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

// Tests for [parseFuncs] function.
func Test_parseFuncs(t *testing.T) {
	t.Run("should return empty funcDecls and funcOrder if astFile has empty Decls", func(t *testing.T) {
		astFile := &ast.File{
			Decls: nil,
		}
		fileSet := token.NewFileSet()
		funcDecls, funcOrder := parseFuncs(astFile, fileSet)
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

		funcDecls, funcOrder := parseFuncs(astFile, fileSet)
		require.Len(t, funcDecls, 1)
		require.Contains(t, funcDecls, "Foo@file.go:4:1")
		require.Equal(t, []string{"Foo@file.go:4:1"}, funcOrder)
	})

	t.Run("should include function literals and prefer the innermost function", func(t *testing.T) {
		src := `package main

	var handler = func() {
		panic("uncovered")
	}

	func outer() {
		handler := func() {
			panic("also uncovered")
		}
		_ = handler
	}`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecls, funcOrder := parseFuncs(astFile, fileSet)
		require.Len(t, funcDecls, 3)
		require.Len(t, funcOrder, 3)

		parsedFile := File{FuncDecls: funcDecls}
		require.Contains(t, parsedFile.LineFunctionName(4), "func@")
		require.Contains(t, parsedFile.LineFunctionName(9), "func@")
		require.Equal(t, "outer@file.go:7:2", parsedFile.LineFunctionName(12))
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

		funcDecls, funcOrder := parseFuncs(astFile, fileSet)
		require.Len(t, funcDecls, 2)
		require.Contains(t, funcDecls, "MyType.Foo@file.go:6:1")
		require.Contains(t, funcDecls, "MyType2.Foo@file.go:10:1")
		require.Equal(t, []string{"MyType.Foo@file.go:6:1", "MyType2.Foo@file.go:10:1"}, funcOrder)
	})

	t.Run("should use physical line numbers when the file has a line directive", func(t *testing.T) {
		src := `package main
//line mapped.go:100
func Foo() {
}
`
		expectedFuncDecl := &FuncDecl{
			Name:  "Foo@file.go:3:1",
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

		funcDecls, funcOrder := parseFuncs(astFile, fileSet)
		require.Equal(t, []string{"Foo@file.go:3:1"}, funcOrder)
		require.Equal(t, expectedFuncDecl, funcDecls["Foo@file.go:3:1"])
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

		funcDecls, funcOrder := parseFuncs(astFile, fileSet)
		require.Len(t, funcDecls, 2)
		require.Len(t, funcOrder, 2)
		require.Contains(t, funcOrder[0], "init@")
		require.Contains(t, funcOrder[1], "init@")
	})

	t.Run("should handle function declaration with nil fileSet", func(t *testing.T) {
		src := `
package main

func Foo() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecls, funcOrder := parseFuncs(astFile, nil)
		require.Empty(t, funcDecls)
		require.Empty(t, funcOrder)
	})

	t.Run("should handle function literal with nil fileSet", func(t *testing.T) {
		src := `
package main

var Foo = func() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecls, funcOrder := parseFuncs(astFile, nil)
		require.Empty(t, funcDecls)
		require.Empty(t, funcOrder)
	})
}

// Tests for [parseFuncDecl] function.
func Test_parseFuncDecl(t *testing.T) {
	t.Run("should return nil if funcDecl is nil", func(t *testing.T) {
		require.Nil(t, parseFuncDecl(nil, token.NewFileSet()))
	})

	t.Run("should return nil if fileSet is nil", func(t *testing.T) {
		require.Nil(t, parseFuncDecl(&ast.FuncDecl{}, nil))
	})

	t.Run("should return new func declaration if funcDecl and fileSet are valid", func(t *testing.T) {
		src := `
package main

func Foo() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecl := astFile.Decls[0].(*ast.FuncDecl)
		result := parseFuncDecl(funcDecl, fileSet)
		require.NotNil(t, result)
		require.Equal(t, "Foo@file.go:4:1", result.Name)
	})

	t.Run("should return new func declaration even if funcDecl.Body is nil", func(t *testing.T) {
		src := `
package main

func Foo()
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcDecl := astFile.Decls[0].(*ast.FuncDecl)
		funcDecl.Body = nil
		result := parseFuncDecl(funcDecl, fileSet)
		require.NotNil(t, result)
		require.Equal(t, "Foo@file.go:4:1", result.Name)
		assert.Equal(t, result.Body.Start, result.Body.End)
	})
}

// Tests for [parseFuncLit] function.
func Test_parseFuncLit(t *testing.T) {
	t.Run("should return nil if funcLit is nil", func(t *testing.T) {
		require.Nil(t, parseFuncLit(nil, token.NewFileSet()))
	})

	t.Run("should return nil if fileSet is nil", func(t *testing.T) {
		require.Nil(t, parseFuncLit(&ast.FuncLit{}, nil))
	})

	t.Run("should return new func declaration if funcLit and fileSet are valid", func(t *testing.T) {
		src := `
package main

var Foo = func() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcLit := astFile.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0].(*ast.FuncLit)
		result := parseFuncLit(funcLit, fileSet)
		require.NotNil(t, result)
		require.Equal(t, "func@file.go:4:11", result.Name)
	})

	t.Run("should return new func literal even if funcLit.Body is nil", func(t *testing.T) {
		src := `
package main

var Foo = func() {}
`
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, "file.go", src, parser.ParseComments)
		require.NoError(t, err)

		funcLit := astFile.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0].(*ast.FuncLit)
		funcLit.Body = nil
		result := parseFuncLit(funcLit, fileSet)
		require.NotNil(t, result)
		require.Equal(t, "func@file.go:4:11", result.Name)
		assert.Equal(t, result.Body.Start, result.Body.End)
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
