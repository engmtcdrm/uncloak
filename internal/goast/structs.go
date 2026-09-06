package goast

import "slices"

type FileFuncDecls struct {
	Path      string
	FuncDecls []*FuncDecl
}

func NewFileFuncDecls(path string, funcDecls []*FuncDecl) *FileFuncDecls {
	return &FileFuncDecls{
		Path:      path,
		FuncDecls: funcDecls,
	}
}

func (ffd *FileFuncDecls) LineFunctionName(line int) string {
	for _, decl := range ffd.FuncDecls {
		if name := decl.LineFunctionName(line); name != "" {
			return name
		}
	}

	return ""
}

type FuncDecl struct {
	Name         string   // Name of the function.
	Lines        []int    // Lines covered by the function.
	ContentLines []string // Lines of code content for the function. index 0 is the function declaration line.
	StartLine    int      // Starting line of the function.
	EndLine      int      // Ending line of the function.
}

// LineFunctionName returns the name of the function that covers the given line.
// If the line is not covered by the function, it returns an empty string.
func (fd *FuncDecl) LineFunctionName(line int) string {
	if slices.Contains(fd.Lines, line) {
		return fd.Name
	}

	return ""
}

// NewFuncDecl creates a new FuncDecl with the given name, start line, and end
// line. It automatically generates the list of lines covered by the function.
func NewFuncDecl(name string, startLine, endLine int, codeLines []string) *FuncDecl {
	lines := make([]int, 0, endLine-startLine+1)
	for line := startLine; line <= endLine; line++ {
		lines = append(lines, line)
	}

	return &FuncDecl{
		Name:         name,
		Lines:        lines,
		ContentLines: codeLines,
		StartLine:    startLine,
		EndLine:      endLine,
	}
}
