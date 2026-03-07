// Package slicemaplint provides analyzers for detecting slice and map issues.
package slicemaplint

import (
	"go/ast"
	"go/types"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/nolint"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for slicemaplint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL060": {
		ID:       "AIL060",
		Name:     "nil-map-write",
		Severity: domain.SeverityCritical,
		Category: domain.CategoryNil,
		Message:  "write to nil map will panic",
		Why: `Writing to a nil map causes a runtime panic. Unlike slices, maps
must be initialized with make() or a map literal before use.

This is a critical bug that will crash your program at runtime.`,
		Fix: `Initialize the map before writing to it using make() or a map literal.`,
		Example: domain.FixExample{
			Bad: `var m map[string]int
m["key"] = 1  // PANIC: assignment to entry in nil map`,
			Good: `m := make(map[string]int)
m["key"] = 1  // OK`,
			Explanation: "Maps must be initialized before writing. Use make(map[K]V) or map[K]V{}.",
		},
		CommonMistakes: []string{
			"WRONG: Declaring with var and forgetting to initialize",
			"WRONG: Assuming map is automatically initialized like a slice",
		},
	},
}

// analyzer implements the slicemaplint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new slicemaplint analyzer.
func New() ports.Analyzer {
	sliceMapAnalyzer := &analyzer{}
	sliceMapAnalyzer.analysis = &analysis.Analyzer{
		Name:     "slicemaplint",
		Doc:      "detects slice and map issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      sliceMapAnalyzer.run,
	}
	return sliceMapAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "slicemaplint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Track potentially nil maps per function
	// Key: *ast.Object (variable), Value: true if potentially nil
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		var body *ast.BlockStmt
		switch funcNode := node.(type) {
		case *ast.FuncDecl:
			if funcNode.Body == nil {
				return
			}
			body = funcNode.Body
		case *ast.FuncLit:
			body = funcNode.Body
		}

		a.checkFunctionBody(pass, body)
	})

	return nil, nil
}

// checkFunctionBody analyzes a function body for nil map writes.
func (a *analyzer) checkFunctionBody(pass *analysis.Pass, body *ast.BlockStmt) {
	// Track which map variables are potentially nil
	// We use a simple approach: if declared with 'var' and no initializer, it's nil
	// If assigned with make() or literal, it's not nil
	nilMaps := make(map[*types.Var]bool)

	ast.Inspect(body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.DeclStmt:
			// Check for var declarations
			genDecl, isGenDecl := stmt.Decl.(*ast.GenDecl)
			if !isGenDecl {
				return true
			}
			for _, spec := range genDecl.Specs {
				valueSpec, isValueSpec := spec.(*ast.ValueSpec)
				if !isValueSpec {
					continue
				}
				// Check if declaring a map without initializer
				for idx, name := range valueSpec.Names {
					obj := pass.TypesInfo.Defs[name]
					if obj == nil {
						continue
					}
					varObj, isVar := obj.(*types.Var)
					if !isVar {
						continue
					}
					// Check if it's a map type
					if !isMapType(varObj.Type()) {
						continue
					}
					// Check if there's an initializer
					if len(valueSpec.Values) > idx {
						// Has initializer - check if it's nil or make/literal
						if !isNilInitializer(valueSpec.Values[idx]) {
							nilMaps[varObj] = false
							continue
						}
					}
					// No initializer or nil initializer - map is nil
					nilMaps[varObj] = true
				}
			}

		case *ast.AssignStmt:
			// Check for assignments that initialize a map
			for idx, lhs := range stmt.Lhs {
				ident, isIdent := lhs.(*ast.Ident)
				if !isIdent {
					continue
				}
				obj := pass.TypesInfo.Uses[ident]
				if obj == nil {
					obj = pass.TypesInfo.Defs[ident]
				}
				if obj == nil {
					continue
				}
				varObj, isVar := obj.(*types.Var)
				if !isVar {
					continue
				}
				if !isMapType(varObj.Type()) {
					continue
				}
				// Check if RHS is make() or literal
				if idx < len(stmt.Rhs) {
					if !isNilInitializer(stmt.Rhs[idx]) {
						nilMaps[varObj] = false
					}
				}
			}

		case *ast.IndexExpr:
			// Check for map write: m[key] = value
			// This is detected in AssignStmt where LHS contains IndexExpr
			// But we need to handle it separately to catch the actual write

		case *ast.IfStmt:
			// Check for nil check pattern: if m == nil { m = make(...) }
			a.checkNilCheckPattern(pass, stmt, nilMaps)
		}

		return true
	})

	// Second pass: find writes to nil maps
	ast.Inspect(body, func(node ast.Node) bool {
		assignStmt, isAssign := node.(*ast.AssignStmt)
		if !isAssign {
			return true
		}

		for _, lhs := range assignStmt.Lhs {
			indexExpr, isIndex := lhs.(*ast.IndexExpr)
			if !isIndex {
				continue
			}

			// Check if indexing into a potentially nil map
			ident, isIdent := indexExpr.X.(*ast.Ident)
			if !isIdent {
				continue
			}

			obj := pass.TypesInfo.Uses[ident]
			if obj == nil {
				continue
			}

			varObj, isVar := obj.(*types.Var)
			if !isVar {
				continue
			}

			if !isMapType(varObj.Type()) {
				continue
			}

			if nilMaps[varObj] {
				diag := Diagnostics["AIL060"]
				nolint.Report(pass, analysis.Diagnostic{
					Pos:      indexExpr.Pos(),
					Category: string(diag.Category),
					Message:  "AIL060: " + diag.Message,
				})
			}
		}

		return true
	})
}

// checkNilCheckPattern detects if m == nil { m = make(...) } pattern.
func (a *analyzer) checkNilCheckPattern(pass *analysis.Pass, ifStmt *ast.IfStmt, nilMaps map[*types.Var]bool) {
	// Check if condition is "m == nil"
	binExpr, isBin := ifStmt.Cond.(*ast.BinaryExpr)
	if !isBin {
		return
	}

	ident, isIdent := binExpr.X.(*ast.Ident)
	if !isIdent {
		return
	}

	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return
	}

	varObj, isVar := obj.(*types.Var)
	if !isVar {
		return
	}

	// Check if the body initializes the map
	ast.Inspect(ifStmt.Body, func(node ast.Node) bool {
		assignStmt, isAssign := node.(*ast.AssignStmt)
		if !isAssign {
			return true
		}

		for idx, lhs := range assignStmt.Lhs {
			lhsIdentifier, isLHSIdent := lhs.(*ast.Ident)
			if !isLHSIdent {
				continue
			}

			lhsObj := pass.TypesInfo.Uses[lhsIdentifier]
			if lhsObj == nil {
				lhsObj = pass.TypesInfo.Defs[lhsIdentifier]
			}

			if lhsObj == varObj && idx < len(assignStmt.Rhs) {
				if !isNilInitializer(assignStmt.Rhs[idx]) {
					nilMaps[varObj] = false
				}
			}
		}

		return true
	})
}

// isMapType checks if a type is a map type.
func isMapType(t types.Type) bool {
	_, isMap := t.Underlying().(*types.Map)
	return isMap
}

// isNilInitializer checks if an expression is nil or doesn't initialize the map.
func isNilInitializer(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "nil"
	case *ast.CallExpr:
		// Check for make()
		if ident, isIdent := e.Fun.(*ast.Ident); isIdent {
			return ident.Name != "make"
		}
		return true
	case *ast.CompositeLit:
		// Map literal - not nil
		return false
	default:
		return true
	}
}
