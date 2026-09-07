package goast

import (
	"go/ast"
	"go/token"
)

// Parse parses the specified Go source file and returns a [*File] containing
// the lines of code and function declarations found in the file.
func Parse(filePath string) (*File, error) {
	return newFile(filePath)
}

// parseFuncDecls extracts all function declarations from the given AST file and
// returns a slice of FuncDecl representing all functions found in the AST file.
func parseFuncDecls(astFile *ast.File, fileSet *token.FileSet) (FuncDecls, []string) {
	funcDecls := make(FuncDecls, len(astFile.Decls))
	funcOrder := make([]string, 0)

	const nl = "\n"
	const cr = "\r\n"

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

		funcDecls[funcDecl.Name.Name] = newFuncDecl(
			funcDecl.Name.Name,
			start,
			end,
		)

		funcOrder = append(funcOrder, funcDecl.Name.Name)
	}

	return funcDecls, funcOrder
}
