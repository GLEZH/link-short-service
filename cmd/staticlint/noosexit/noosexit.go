// Package noosexit detects direct os.Exit calls in the main function
package noosexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer prohibits direct os.Exit calls in the main function of package main
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "prohibit direct os.Exit calls in the main function of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv != nil || function.Name.Name != "main" {
				continue
			}

			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				if isOSExit(pass, call) {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main is prohibited")
				}
				return true
			})
		}
	}

	return nil, nil
}

func isOSExit(pass *analysis.Pass, call *ast.CallExpr) bool {
	var object types.Object

	switch function := call.Fun.(type) {
	case *ast.Ident:
		object = pass.TypesInfo.Uses[function]
	case *ast.SelectorExpr:
		object = pass.TypesInfo.Uses[function.Sel]
	default:
		return false
	}

	if object == nil || object.Pkg() == nil {
		return false
	}

	_, ok := object.(*types.Func)
	return ok && object.Pkg().Path() == "os" && object.Name() == "Exit"
}
