// Package exitcheck defines an Analyzer that reports os.Exit function usage.
package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer describes exitcheck analyzer and its options.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "checker helps determine if os.exit exists",
	Run:  Run,
}

// Run applies the analyzer to a package.
// It returns an error if the analyzer failed.
func Run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if x, ok := node.(*ast.CallExpr); ok {
				if s, ok := x.Fun.(*ast.SelectorExpr); ok {
					if s.Sel.Name == "Exit" {
						pass.Reportf(s.Pos(), "using os.Exit is prohibbited")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
