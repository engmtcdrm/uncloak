package goast

import (
	"maps"
)

// File represents a Go source file, including its path, lines of code, and
// function declarations.
type File struct {
	Path      string
	Lines     []string
	FuncDecls FuncDecls
	FuncOrder []string
}

// LineContent returns the content of the given line in the file. If the line
// number is out of range, it returns an empty string.
func (f File) LineContent(line int) string {
	if line <= 0 || line > len(f.Lines) {
		return ""
	}

	return f.Lines[line-1]
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

// FuncBody represents the body of a function, including the lines it covers and
// its start and end lines.
type FuncBody struct {
	Lines []int // Lines covered by the function body.
	Start int   // Starting line of the function body.
	End   int   // Ending line of the function body.
}

// newFuncBody creates a new [FuncBody] with the given start and end lines, and
// generates the list of lines covered by the function body.
func newFuncBody(start, end int) FuncBody {
	lines := make([]int, 0, end-start+1)
	for line := start; line <= end; line++ {
		lines = append(lines, line)
	}

	return FuncBody{
		Lines: lines,
		Start: start,
		End:   end,
	}
}

// FuncDecl represents a function declaration within a Go source file, including
// its name, the lines it covers, and its start and end lines.
type FuncDecl struct {
	Name  string   // Name of the function.
	Lines []int    // Lines covered by the function.
	Start int      // Starting line of the function.
	End   int      // Ending line of the function.
	Body  FuncBody // The body of the function.
}

// newFuncDecl creates a new [*FuncDecl] with the given name, start line, and
// end line, and the function body. It automatically generates the list of lines
// covered by the function.
func newFuncDecl(name string, start, end int, body FuncBody) *FuncDecl {
	lines := make([]int, 0, end-start+1)
	for line := start; line <= end; line++ {
		lines = append(lines, line)
	}

	return &FuncDecl{
		Name:  name,
		Lines: lines,
		Start: start,
		End:   end,
		Body:  body,
	}
}

// LineFunctionName returns the name of the function that covers the given line.
// If the line is not covered by the function, it returns an empty string.
func (fd FuncDecl) LineFunctionName(line int) string {
	if fd.Start > line || fd.End < line {
		return ""
	}

	return fd.Name
}

// FuncDecls represents a collection of [FuncDecl], mapped by their names.
type FuncDecls map[string]*FuncDecl

// Names returns a slice of all function names in the collection.
func (f FuncDecls) Names() []string {
	names := make([]string, 0, len(f))

	for name := range maps.Keys(f) {
		names = append(names, name)
	}

	return names
}
