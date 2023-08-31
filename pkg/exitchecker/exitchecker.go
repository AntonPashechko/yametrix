// Package exitchecker анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
package exitchecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for calling os.Exit in main/main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name != "main" { //Проверка на main пакет
					return false
				}
			case *ast.FuncDecl:
				if x.Name.Name != "main" { //Проверка на main функцию
					return false
				}
			case *ast.CallExpr:
				if s, ok := x.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := s.X.(*ast.Ident); ok {
						if ident.Name == "os" && s.Sel.Name == "Exit" {
							pass.Reportf(ident.NamePos, "os.Exit called in main/main")
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
