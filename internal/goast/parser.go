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

	funcDecls, funcOrder := parseFuncs(astFile, fileSet)

	return &File{
		Path:      filePath,
		Lines:     strings.Split(string(src), "\n"),
		FuncDecls: funcDecls,
		FuncOrder: funcOrder,
	}, nil
}

// parseFuncs parses the AST of a Go source file and extracts function
// declarations and their order of appearance. It returns a map of function
// names to [FuncDecl] and a slice of function names in the order they appear in
// the source file. This function will parse both function declarations and
// function literals.
func parseFuncs(astFile *ast.File, fileSet *token.FileSet) (FuncDecls, []string) {
	funcDecls := make(FuncDecls, len(astFile.Decls))
	funcOrder := make([]string, 0)

	ast.Inspect(astFile, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			newDecl := parseFuncDecl(n, fileSet)
			if newDecl == nil {
				return true
			}

			funcDecls[newDecl.Name] = newDecl
			funcOrder = append(funcOrder, newDecl.Name)
		case *ast.FuncLit:
			newDecl := parseFuncLit(n, fileSet)
			if newDecl == nil {
				return true
			}

			funcDecls[newDecl.Name] = newDecl
			funcOrder = append(funcOrder, newDecl.Name)
		}

		return true
	})

	return funcDecls, funcOrder
}

// parseFuncDecl parses an AST function declaration and returns a [FuncDecl]
// representing it. If the input is nil, it returns nil.
func parseFuncDecl(funcDecl *ast.FuncDecl, fileSet *token.FileSet) *FuncDecl {
	if funcDecl == nil || fileSet == nil {
		return nil
	}

	start := fileSet.PositionFor(funcDecl.Pos(), false)
	end := fileSet.PositionFor(funcDecl.End(), false).Line
	bodyStart := start.Line
	bodyEnd := start.Line

	if funcDecl.Body != nil {
		bodyStart = fileSet.PositionFor(funcDecl.Body.Pos(), false).Line
		bodyEnd = fileSet.PositionFor(funcDecl.Body.End(), false).Line
	}

	name := funcDecl.Name.Name + "@" + start.String()

	if recv := receiverTypeName(funcDecl); recv != "" {
		name = recv + "." + name
	}

	return newFuncDecl(
		name,
		start.Line,
		end,
		newFuncBody(bodyStart, bodyEnd),
	)
}

// parseFuncLit parses an AST function literal and returns a [FuncDecl]
// representing it. If the input is nil, it returns nil.
func parseFuncLit(funcLit *ast.FuncLit, fileSet *token.FileSet) *FuncDecl {
	if funcLit == nil || fileSet == nil {
		return nil
	}

	start := fileSet.PositionFor(funcLit.Pos(), false)
	end := start.Line
	bodyStart := start.Line
	bodyEnd := start.Line

	if funcLit.Body != nil {
		end = fileSet.PositionFor(funcLit.End(), false).Line
		bodyStart = fileSet.PositionFor(funcLit.Body.Pos(), false).Line
		bodyEnd = fileSet.PositionFor(funcLit.Body.End(), false).Line
	}

	name := "func@" + start.String()

	newDecl := newFuncLit(
		name,
		start.Line,
		end,
		newFuncBody(bodyStart, bodyEnd),
	)

	return newDecl
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
