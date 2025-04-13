// Package analyzer provides static analysis tools for the URL shortener service.
package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// OSExitAnalyzer проверяет, что в функции main пакета main нет прямого вызова os.Exit.
var OSExitAnalyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "check for direct calls to os.Exit in the main function of the main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Проверяем, что анализ выполняется для пакета main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Используем inspector для обхода AST
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Фильтр для функции main
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Name.Name == "main" {
			// Проверяем каждое выражение в теле функции main
			for _, stmt := range fn.Body.List {
				if expr, ok := stmt.(*ast.ExprStmt); ok {
					if call, ok := expr.X.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok {
								if ident.Name == "os" && sel.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "direct call to os.Exit in main function")
								}
							}
						}
					}
				}
			}
		}
	})

	return nil, nil
}
