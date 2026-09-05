package gocover

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strings"
)

type FuncDecls map[string][]FuncDecl

type FuncDecl struct {
	Name      string
	StartLine int
	EndLine   int
}

func NewFuncDecl(name string, startLine, endLine int) FuncDecl {
	return FuncDecl{
		Name:      name,
		StartLine: startLine,
		EndLine:   endLine,
	}
}

func IsFunction(file string, line int, funcDecls FuncDecls) bool {
	funcDecl, ok := funcDecls[file]
	if !ok {
		return false
	}

	for _, fd := range funcDecl {
		if line >= fd.StartLine && line <= fd.EndLine {
			return true
		}
	}

	return false
}

func ParseGoFiles(files []string) (FuncDecls, error) {
	funcList := FuncDecls{}

	for _, file := range files {
		fileFuncDecl, err := ParseGoFile(file)
		if err != nil {
			return nil, err
		}

		if fileFuncDecl == nil {
			continue
		}

		funcList[file] = append(funcList[file], fileFuncDecl...)

	}

	return funcList, nil
}

func ParseGoFile(filePath string) ([]FuncDecl, error) {
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

	fileSet := token.NewFileSet()
	astFile, err := goparser.ParseFile(fileSet, filePath, nil, 4)
	if err != nil {
		return nil, err
	}

	return getFuncDecls(astFile, fileSet), nil
}

func getFuncDecls(astFile *ast.File, fileSet *token.FileSet) []FuncDecl {
	funcDecls := []FuncDecl{}

	for _, decl := range astFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)

		if !ok {
			continue
		}

		start := fileSet.Position(funcDecl.Pos()).Line
		end := start

		if funcDecl.Body != nil && len(funcDecl.Body.List) > 0 {
			// start = fileSet.Position(funcDecl.Body.List[0].Pos()).Line
			end = fileSet.Position(funcDecl.Body.List[len(funcDecl.Body.List)-1].End()).Line
		}

		funcDecls = append(funcDecls, NewFuncDecl(
			funcDecl.Name.Name,
			start,
			end,
		))
	}

	return funcDecls
}
