package goast

import (
	"go/parser"
	"go/token"
	"maps"
	"os"
	"strings"
)

// File represents a Go source file, including its path, lines of code, and
// function declarations.
type File struct {
	Path      string
	Lines     []string
	FuncDecls FuncDecls
	FuncOrder []string
}

// newFile creates a new [*File] instance for the given file path, parsing its
// content and extracting function declarations. It returns an error if the file
// cannot be read or parsed.
func newFile(path string) (*File, error) {
	if !strings.HasSuffix(path, ".go") {
		return nil, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nil
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, path, src, 4)
	if err != nil {
		return nil, err
	}

	funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)

	return &File{
		Path:      path,
		Lines:     strings.Split(string(src), "\n"),
		FuncDecls: funcDecls,
		FuncOrder: funcOrder,
	}, nil
}

// LineFunctionName returns the name of the function that covers the given line
// in the file.
func (f File) LineFunctionName(line int) string {
	for _, decl := range f.FuncDecls {
		if name := decl.LineFunctionName(line); name != "" {
			return name
		}
	}

	return ""
}

// LineContent returns the content of the given line in the file. If the line
// number is out of range, it returns an empty string.
func (f File) LineContent(line int) string {
	if line > 0 && line <= len(f.Lines) {
		return f.Lines[line-1]
	}

	return ""
}

// FuncDecl represents a function declaration within a Go source file, including
// its name, the lines it covers, and its start and end lines.
type FuncDecl struct {
	Name      string // Name of the function.
	Lines     []int  // Lines covered by the function.
	StartLine int    // Starting line of the function.
	EndLine   int    // Ending line of the function.
}

// newFuncDecl creates a new [*FuncDecl] with the given name, start line, and
// end line. It automatically generates the list of lines covered by the
// function.
func newFuncDecl(name string, startLine, endLine int) *FuncDecl {
	lines := make([]int, 0, endLine-startLine+1)
	for line := startLine; line <= endLine; line++ {
		lines = append(lines, line)
	}

	return &FuncDecl{
		Name:      name,
		Lines:     lines,
		StartLine: startLine,
		EndLine:   endLine,
	}
}

// LineFunctionName returns the name of the function that covers the given line.
// If the line is not covered by the function, it returns an empty string.
func (fd FuncDecl) LineFunctionName(line int) string {
	if fd.StartLine <= line && line <= fd.EndLine {
		return fd.Name
	}

	return ""
}

// FuncDecls represents a collection of [FuncDecl], mapped by their names.
type FuncDecls map[string]*FuncDecl

// Names returns a slice of all function names in the collection.
func (f FuncDecls) Names() []string {
	names := make([]string, 0, len(f))

	k := maps.Keys(f)

	for name := range k {
		names = append(names, name)
	}

	return names
}
