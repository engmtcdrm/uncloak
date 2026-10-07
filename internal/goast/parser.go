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

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, filePath, src, parser.ParseComments)
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
// returns a slice of [FuncDecl] representing all functions found in the AST
// file.
func parseFuncDecls(astFile *ast.File, fileSet *token.FileSet) (FuncDecls, []string) {
	funcDecls := make(FuncDecls, len(astFile.Decls))
	funcOrder := make([]string, 0)

	for _, decl := range astFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		start := fileSet.PositionFor(funcDecl.Pos(), false).Line
		bodyStart := start
		bodyEnd := start
		end := start

		if funcDecl.Body != nil {
			bodyStart = fileSet.PositionFor(funcDecl.Body.Pos(), false).Line
			bodyEnd = fileSet.PositionFor(funcDecl.Body.End(), false).Line
			end = fileSet.PositionFor(funcDecl.End(), false).Line
		}

		name := funcDecl.Name.Name
		if name == "init" && funcDecl.Recv == nil {
			name += "@" + fileSet.PositionFor(funcDecl.Pos(), false).String()
		}

		if recv := receiverTypeName(funcDecl); recv != "" {
			name = recv + "." + name
		}

		funcDecls[name] = newFuncDecl(
			name,
			start,
			end,
			newFuncBody(bodyStart, bodyEnd),
		)

		funcOrder = append(funcOrder, name)
	}

	return funcDecls, funcOrder
}

// receiverTypeName returns the type name of the receiver for the given function
// declaration. If the function has no receiver, an empty string is returned.
func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	return receiverBaseTypeName(fn.Recv.List[0].Type)
}

// receiverBaseTypeName returns the base type name of the given receiver
// expression. It handles pointer receivers, indexed types, and generic types
// recursively.
func receiverBaseTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}

		return receiverBaseTypeName(t.X)
	case *ast.IndexExpr:
		return receiverBaseTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverBaseTypeName(t.X)
	default:
		return ""
	}
}
