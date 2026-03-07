// Package slicemaplint provides analyzers for detecting slice and map issues.
package slicemaplint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
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
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

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
	nilMaps := make(map[*types.Var]bool)

	// First pass: track nil maps
	ast.Inspect(body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.DeclStmt:
			a.trackDeclStmt(pass, stmt, nilMaps)
		case *ast.AssignStmt:
			a.trackAssignStmt(pass, stmt, nilMaps)
		case *ast.IfStmt:
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
		a.checkNilMapWrite(pass, assignStmt, nilMaps)
		return true
	})
}

// trackDeclStmt tracks map variable declarations.
func (a *analyzer) trackDeclStmt(pass *analysis.Pass, stmt *ast.DeclStmt, nilMaps map[*types.Var]bool) {
	genDecl, isGenDecl := stmt.Decl.(*ast.GenDecl)
	if !isGenDecl {
		return
	}
	for _, spec := range genDecl.Specs {
		valueSpec, isValueSpec := spec.(*ast.ValueSpec)
		if !isValueSpec {
			continue
		}
		a.trackValueSpec(pass, valueSpec, nilMaps)
	}
}

// trackValueSpec tracks map variables in a value spec.
func (a *analyzer) trackValueSpec(pass *analysis.Pass, valueSpec *ast.ValueSpec, nilMaps map[*types.Var]bool) {
	for idx, name := range valueSpec.Names {
		obj := pass.TypesInfo.Defs[name]
		if obj == nil {
			continue
		}
		varObj, isVar := obj.(*types.Var)
		if !isVar || !isMapType(varObj.Type()) {
			continue
		}
		// Check if there's a non-nil initializer
		if len(valueSpec.Values) > idx && !isNilInitializer(valueSpec.Values[idx]) {
			nilMaps[varObj] = false
			continue
		}
		nilMaps[varObj] = true
	}
}

// trackAssignStmt tracks map assignments.
func (a *analyzer) trackAssignStmt(pass *analysis.Pass, stmt *ast.AssignStmt, nilMaps map[*types.Var]bool) {
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
		if !isVar || !isMapType(varObj.Type()) {
			continue
		}
		if idx < len(stmt.Rhs) && !isNilInitializer(stmt.Rhs[idx]) {
			nilMaps[varObj] = false
		}
	}
}

// checkNilMapWrite checks if an assignment writes to a nil map.
func (a *analyzer) checkNilMapWrite(pass *analysis.Pass, assignStmt *ast.AssignStmt, nilMaps map[*types.Var]bool) {
	for _, lhs := range assignStmt.Lhs {
		indexExpr, isIndex := lhs.(*ast.IndexExpr)
		if !isIndex {
			continue
		}
		ident, isIdent := indexExpr.X.(*ast.Ident)
		if !isIdent {
			continue
		}
		obj := pass.TypesInfo.Uses[ident]
		if obj == nil {
			continue
		}
		varObj, isVar := obj.(*types.Var)
		if !isVar || !isMapType(varObj.Type()) {
			continue
		}
		if nilMaps[varObj] {
			diag := Diagnostics["AIL060"]
			domain.Report(pass, analysis.Diagnostic{
				Pos:      indexExpr.Pos(),
				Category: string(diag.Category),
				Message:  "AIL060: " + diag.Message,
			})
		}
	}
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
