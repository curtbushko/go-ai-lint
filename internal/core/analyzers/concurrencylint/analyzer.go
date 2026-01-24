// Package concurrencylint provides analyzers for detecting concurrency issues.
package concurrencylint

import (
	"go/ast"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for concurrencylint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL080": {
		ID:       "AIL080",
		Name:     "waitgroup-done-not-deferred",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryConcurrency,
		Message:  "wg.Done() should be deferred",
		Why: `When wg.Done() is not deferred, it may not be called if the goroutine
panics or returns early. This causes wg.Wait() to block forever, leading
to goroutine leaks and deadlocks.`,
		Fix: `Use defer wg.Done() at the start of the goroutine. This ensures
Done() is called regardless of how the goroutine exits.`,
		Example: domain.FixExample{
			Bad: `go func() {
    doWork()
    wg.Done() // May not run if doWork panics
}()`,
			Good: `go func() {
    defer wg.Done()
    doWork() // Done() always runs
}()`,
			Explanation: "Defer guarantees wg.Done() runs even on panic or early return.",
		},
		CommonMistakes: []string{
			"WRONG: Putting wg.Done() at the end without defer",
			"WRONG: Having multiple wg.Done() calls in different branches",
		},
	},
	"AIL082": {
		ID:       "AIL082",
		Name:     "select-only-default",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryConcurrency,
		Message:  "select with only default case is useless",
		Why: `A select statement with only a default case executes immediately
without any channel operations. The select is redundant and should
be removed for clarity.`,
		Fix: `Remove the select statement and execute the code directly,
or add channel cases if concurrent behavior is intended.`,
		Example: domain.FixExample{
			Bad: `select {
default:
    doWork()
}`,
			Good: `doWork()`,
			Explanation: "Without channel cases, select adds no value.",
		},
		CommonMistakes: []string{
			"WRONG: Confusing 'select with default' with non-blocking receive",
			"WRONG: Using empty select {} thinking it's the same",
		},
	},
}

// analyzer implements the concurrencylint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new concurrencylint analyzer.
func New() ports.Analyzer {
	concurrencyAnalyzer := &analyzer{}
	concurrencyAnalyzer.analysis = &analysis.Analyzer{
		Name:     "concurrencylint",
		Doc:      "detects concurrency issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      concurrencyAnalyzer.run,
	}
	return concurrencyAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "concurrencylint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Check for AIL080: wg.Done() not deferred in goroutine
	goStmtFilter := []ast.Node{
		(*ast.GoStmt)(nil),
	}

	insp.Preorder(goStmtFilter, func(node ast.Node) {
		goStmt := node.(*ast.GoStmt)
		a.checkWaitGroupDone(pass, goStmt)
	})

	// Check for AIL082: select with only default case
	selectStmtFilter := []ast.Node{
		(*ast.SelectStmt)(nil),
	}

	insp.Preorder(selectStmtFilter, func(node ast.Node) {
		selectStmt := node.(*ast.SelectStmt)
		a.checkSelectOnlyDefault(pass, selectStmt)
	})

	return nil, nil
}

// checkWaitGroupDone checks for AIL080: wg.Done() not deferred in goroutine.
func (a *analyzer) checkWaitGroupDone(pass *analysis.Pass, goStmt *ast.GoStmt) {
	// Get the function literal from the go statement
	funcLit, isFuncLit := goStmt.Call.Fun.(*ast.FuncLit)
	if !isFuncLit {
		return
	}

	// Find all Done() calls that are NOT inside defer statements
	ast.Inspect(funcLit.Body, func(node ast.Node) bool {
		// Skip defer statements - deferred Done() is correct
		if _, isDefer := node.(*ast.DeferStmt); isDefer {
			return false
		}

		exprStmt, isExprStmt := node.(*ast.ExprStmt)
		if !isExprStmt {
			return true
		}

		callExpr, isCall := exprStmt.X.(*ast.CallExpr)
		if !isCall {
			return true
		}

		if a.isWaitGroupDone(callExpr) {
			diag := Diagnostics["AIL080"]
			pass.Report(analysis.Diagnostic{
				Pos:      callExpr.Pos(),
				Category: string(diag.Category),
				Message:  "AIL080: " + diag.Message,
			})
		}

		return true
	})
}

// isWaitGroupDone checks if a call is wg.Done() or similar.
func (a *analyzer) isWaitGroupDone(call *ast.CallExpr) bool {
	selExpr, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return false
	}

	// Check if method is "Done"
	return selExpr.Sel.Name == "Done"
}

// checkSelectOnlyDefault checks for AIL082: select with only default case.
func (a *analyzer) checkSelectOnlyDefault(pass *analysis.Pass, selectStmt *ast.SelectStmt) {
	// Empty select {} is a valid pattern (blocks forever), don't flag
	if selectStmt.Body == nil || len(selectStmt.Body.List) == 0 {
		return
	}

	// Check if there's exactly one case and it's the default
	if len(selectStmt.Body.List) != 1 {
		return
	}

	commClause, isComm := selectStmt.Body.List[0].(*ast.CommClause)
	if !isComm {
		return
	}

	// commClause.Comm == nil means it's the default case
	if commClause.Comm == nil {
		diag := Diagnostics["AIL082"]
		pass.Report(analysis.Diagnostic{
			Pos:      selectStmt.Pos(),
			Category: string(diag.Category),
			Message:  "AIL082: " + diag.Message,
		})
	}
}
