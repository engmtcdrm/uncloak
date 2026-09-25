package goast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// Parse parses the specified Go source file and returns a [*File] containing
// the lines of code and function declarations found in the file.
func Parse(filePath string) (*File, error) {
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
	astFile, err := parser.ParseFile(fileSet, filePath, src, 4)
	if err != nil {
		return nil, err
	}

	funcDecls, funcOrder := parseFuncDecls(astFile, fileSet)

	return &File{
		Path:      filePath,
		Lines:     strings.Split(string(src), "\n"),
		FuncDecls: funcDecls,
		FuncOrder: funcOrder,
	}, nil
}

// parseFuncDecls extracts all function declarations from the given AST file and
// returns a slice of FuncDecl representing all functions found in the AST file.
func parseFuncDecls(astFile *ast.File, fileSet *token.FileSet) (FuncDecls, []string) {
	funcDecls := make(FuncDecls, len(astFile.Decls))
	funcOrder := make([]string, 0)

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

		name := funcDecl.Name.Name
		if recv := receiverTypeName(funcDecl); recv != "" {
			name = recv + "." + name
		}

		funcDecls[name] = newFuncDecl(
			name,
			start,
			end,
		)

		funcOrder = append(funcOrder, name)
	}

	return funcDecls, funcOrder
}

// receiverTypeName returns the type name of the receiver for the given function
// declaration.
func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	switch t := fn.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	return ""
}
