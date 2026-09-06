package goast

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strings"
)

// ParseGoFile parses the specified Go source file and returns a list of
// function declarations found in the file.
func ParseGoFile(filePath string) (*FileFuncDecls, error) {
	if !strings.HasSuffix(filePath, ".go") {
		return nil, nil
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, nil
	}

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	astFile, err := goparser.ParseFile(fileSet, filePath, src, 4)
	if err != nil {
		return nil, err
	}

	declarations := getFuncDecls(astFile, fileSet, src)

	return NewFileFuncDecls(filePath, declarations), nil
}

// getFuncDecls extracts all function declarations from the given AST file and
// returns a slice of FuncDecl representing all functions found in the AST file.
func getFuncDecls(astFile *ast.File, fileSet *token.FileSet, src []byte) []*FuncDecl {
	funcDecls := []*FuncDecl{}

	const nl = "\n"
	const cr = "\r\n"

	lines := strings.Split(strings.ReplaceAll(string(src), cr, nl), nl)

	for _, decl := range astFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		start := fileSet.Position(funcDecl.Pos()).Line
		end := start

		if funcDecl.Body != nil && len(funcDecl.Body.List) > 0 {
			end = fileSet.Position(funcDecl.Body.List[len(funcDecl.Body.List)-1].End()).Line + 1
		}

		codeLines := make([]string, 0, end-start+1)
		if start >= 1 && end <= len(lines) && start <= end {
			codeLines = append(codeLines, lines[start-1:end]...)
		}

		funcDecls = append(funcDecls, NewFuncDecl(
			funcDecl.Name.Name,
			start,
			end,
			codeLines,
		))
	}

	return funcDecls
}
