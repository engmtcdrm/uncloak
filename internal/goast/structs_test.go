package goast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [newFile] function.
func Test_newFile(t *testing.T) {
}

// Tests for [File.LineContent] function.
func Test_File_LineContent(t *testing.T) {
	const lineOutOfRange = 8
	const lineInRange = 2

	file := &File{
		Lines: []string{
			"line 1",
			"line 2",
			"line 3",
		},
	}

	t.Run("should return an empty string if line is out of range", func(t *testing.T) {
		result := file.LineContent(lineOutOfRange)
		require.Empty(t, result)
	})

	t.Run("should return the correct line content if line is in range", func(t *testing.T) {
		result := file.LineContent(lineInRange)
		require.Equal(t, "line 2", result)
	})
}

// Tests for [File.LineFunctionName] function.
func Test_File_LineFunctionName(t *testing.T) {
	const funcName = "myFunc"
	const lineOutOfRange = 10
	const lineInRange = 6

	file := &File{
		FuncDecls: FuncDecls{
			funcName: newFuncDecl(funcName, 5, 8),
		},
	}

	t.Run("should return empty string if the line is not part of any function", func(t *testing.T) {
		result := file.LineFunctionName(lineOutOfRange)
		require.Empty(t, result)
	})

	t.Run("should return the function name if the line is part of a function", func(t *testing.T) {
		result := file.LineFunctionName(lineInRange)
		require.Equal(t, funcName, result)
	})
}

// Tests for [newFuncDecl] function.
func Test_newFuncDecl(t *testing.T) {
	t.Run("should panic if StartLine is greater than EndLine", func(t *testing.T) {
		require.Panics(t, func() { newFuncDecl("myFunc", 10, 5) })
	})

	t.Run("should return a valid FuncDecl when StartLine is less than or equal to EndLine", func(t *testing.T) {
		fd := newFuncDecl("myFunc", 5, 10)
		require.Equal(t, "myFunc", fd.Name)
		require.Equal(t, 5, fd.StartLine)
		require.Equal(t, 10, fd.EndLine)
		require.Equal(t, []int{5, 6, 7, 8, 9, 10}, fd.Lines)
	})
}

// Tests for [FuncDecl.LineFunctionName] function.
func Test_FuncDecl_LineFunctionName(t *testing.T) {
	const funcName = "myFunc"
	const lineOutOfRange = 10
	const lineInRange = 6

	fd := newFuncDecl(funcName, 5, 8)

	t.Run("should return empty string if the line is not part of the function", func(t *testing.T) {
		result := fd.LineFunctionName(lineOutOfRange)
		require.Empty(t, result)
	})

	t.Run("should return the function name if the line is part of the function", func(t *testing.T) {
		result := fd.LineFunctionName(lineInRange)
		require.Equal(t, funcName, result)
	})
}

// Tests for [FuncDecls.Names] function.
func Test_FuncDecls_Names(t *testing.T) {
	t.Run("should return empty slice if no names exist", func(t *testing.T) {
		fd := FuncDecls{}
		names := fd.Names()
		require.Empty(t, names)
	})

	t.Run("should return all function names", func(t *testing.T) {
		fd := FuncDecls{}

		expectedFuncs := []string{"Func1", "Func2"}

		for _, name := range expectedFuncs {
			fd[name] = &FuncDecl{Name: name}
		}

		names := fd.Names()
		require.ElementsMatch(t, names, expectedFuncs)
	})
}
