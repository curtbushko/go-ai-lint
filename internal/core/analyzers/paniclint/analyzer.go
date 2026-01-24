// Package paniclint provides analyzers for detecting panic/recover issues.
package paniclint

import (
	"go/ast"
	"strings"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for paniclint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL090": {
		ID:       "AIL090",
		Name:     "panic-in-library",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryPanic,
		Message:  "panic in library code",
		Why: `Libraries should not panic. Panics crash the entire program and
cannot be handled gracefully by callers. Only the caller should decide
whether to panic or handle errors gracefully.`,
		Fix: `Return an error instead of panicking. If you need a Must* variant
for convenience, keep it separate and document that it panics.`,
		Example: domain.FixExample{
			Bad: `func ParseConfig(data []byte) Config {
    if len(data) == 0 {
        panic("empty config")
    }
    return Config{}
}`,
			Good: `func ParseConfig(data []byte) (Config, error) {
    if len(data) == 0 {
        return Config{}, errors.New("empty config")
    }
    return Config{}, nil
}`,
			Explanation: "Returning errors lets callers decide how to handle failures.",
		},
		CommonMistakes: []string{
			"WRONG: Panicking on invalid input",
			"WRONG: Using panic for control flow",
		},
	},
	"AIL091": {
		ID:       "AIL091",
		Name:     "empty-recover",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryPanic,
		Message:  "recover() result discarded",
		Why: `Calling recover() without using its result silently swallows panics.
This hides bugs and makes debugging extremely difficult. If a panic
occurs, it will be completely invisible.`,
		Fix: `At minimum, log the recovered value. Consider whether the panic
should be re-raised or converted to an error.`,
		Example: domain.FixExample{
			Bad: `defer func() {
    recover() // Panic silently swallowed
}()`,
			Good: `defer func() {
    if r := recover(); r != nil {
        log.Printf("recovered from panic: %v", r)
    }
}()`,
			Explanation: "Always use the recover value - at least log it for debugging.",
		},
		CommonMistakes: []string{
			"WRONG: Using recover() just to prevent crashes",
			"WRONG: Assigning recover() to _ (blank identifier)",
		},
	},
}

// analyzer implements the paniclint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new paniclint analyzer.
func New() ports.Analyzer {
	panicAnalyzer := &analyzer{}
	panicAnalyzer.analysis = &analysis.Analyzer{
		Name:     "paniclint",
		Doc:      "detects panic/recover issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      panicAnalyzer.run,
	}
	return panicAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "paniclint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Skip main package - panic is acceptable there
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}

	// Track current function name for Must* exception
	var currentFuncName string
	// Track if we're inside a recover handler (if r := recover(); r != nil { ... })
	inRecoverHandler := false

	// Check for panic and recover issues
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
		(*ast.DeferStmt)(nil),
		(*ast.IfStmt)(nil),
	}

	insp.Nodes(nodeFilter, func(node ast.Node, push bool) bool {
		switch stmt := node.(type) {
		case *ast.FuncDecl:
			if push {
				currentFuncName = stmt.Name.Name
			} else {
				currentFuncName = ""
			}
		case *ast.IfStmt:
			// Check if this is a recover handler: if r := recover(); r != nil
			if push && a.isRecoverHandler(stmt) {
				inRecoverHandler = true
			} else if !push && a.isRecoverHandler(stmt) {
				inRecoverHandler = false
			}
		case *ast.CallExpr:
			if push && !inRecoverHandler {
				a.checkPanicCall(pass, stmt, currentFuncName)
			}
		case *ast.DeferStmt:
			if push {
				a.checkDeferredRecover(pass, stmt)
			}
		}
		return true
	})

	return nil, nil
}

// isRecoverHandler checks if an if statement is a recover handler.
func (a *analyzer) isRecoverHandler(ifStmt *ast.IfStmt) bool {
	// Check for if r := recover(); r != nil pattern
	if ifStmt.Init == nil {
		return false
	}

	assignStmt, isAssign := ifStmt.Init.(*ast.AssignStmt)
	if !isAssign || len(assignStmt.Rhs) != 1 {
		return false
	}

	return a.isRecoverCall(assignStmt.Rhs[0])
}

// checkPanicCall checks for AIL090: panic in library code.
func (a *analyzer) checkPanicCall(pass *analysis.Pass, call *ast.CallExpr, funcName string) {
	ident, isIdent := call.Fun.(*ast.Ident)
	if !isIdent || ident.Name != "panic" {
		return
	}

	// Allow panic in Must* functions
	if strings.HasPrefix(funcName, "Must") {
		return
	}

	diag := Diagnostics["AIL090"]
	pass.Report(analysis.Diagnostic{
		Pos:      call.Pos(),
		Category: string(diag.Category),
		Message:  "AIL090: " + diag.Message,
	})
}

// checkDeferredRecover checks for AIL091: empty recover.
func (a *analyzer) checkDeferredRecover(pass *analysis.Pass, deferStmt *ast.DeferStmt) {
	// Look for defer func() { recover() }() pattern
	funcLit, isFuncLit := deferStmt.Call.Fun.(*ast.FuncLit)
	if !isFuncLit {
		return
	}

	ast.Inspect(funcLit.Body, func(node ast.Node) bool {
		// Check for standalone recover() call (expression statement)
		exprStmt, isExprStmt := node.(*ast.ExprStmt)
		if isExprStmt {
			if a.isRecoverCall(exprStmt.X) {
				a.reportEmptyRecover(pass, exprStmt.X.(*ast.CallExpr))
				return false
			}
		}

		// Check for _ = recover() pattern
		assignStmt, isAssign := node.(*ast.AssignStmt)
		if isAssign && len(assignStmt.Lhs) == 1 && len(assignStmt.Rhs) == 1 {
			if a.isRecoverCall(assignStmt.Rhs[0]) {
				// Check if LHS is blank identifier
				if ident, isIdent := assignStmt.Lhs[0].(*ast.Ident); isIdent && ident.Name == "_" {
					a.reportEmptyRecover(pass, assignStmt.Rhs[0].(*ast.CallExpr))
					return false
				}
			}
		}

		return true
	})
}

// isRecoverCall checks if an expression is a recover() call.
func (a *analyzer) isRecoverCall(expr ast.Expr) bool {
	call, isCall := expr.(*ast.CallExpr)
	if !isCall {
		return false
	}

	ident, isIdent := call.Fun.(*ast.Ident)
	return isIdent && ident.Name == "recover"
}

// reportEmptyRecover reports AIL091.
func (a *analyzer) reportEmptyRecover(pass *analysis.Pass, call *ast.CallExpr) {
	diag := Diagnostics["AIL091"]
	pass.Report(analysis.Diagnostic{
		Pos:      call.Pos(),
		Category: string(diag.Category),
		Message:  "AIL091: " + diag.Message,
	})
}
