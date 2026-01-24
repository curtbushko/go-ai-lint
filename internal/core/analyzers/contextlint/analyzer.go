// Package contextlint provides analyzers for detecting context misuse.
package contextlint

import (
	"go/ast"
	"strings"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for contextlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL010": {
		ID:       "AIL010",
		Name:     "context-todo-production",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryContext,
		Message:  "context.TODO() used in non-test code",
		Why: `context.TODO() is intended as a placeholder during development when
it's unclear which context to use. It signals technical debt and should
not appear in production code.

Using TODO() bypasses proper context propagation, making it impossible
to cancel operations or set deadlines from calling code.`,
		Fix: `Pass context from the caller or use context.Background() if this is
truly a top-level operation that should not be cancellable.`,
		Example: domain.FixExample{
			Bad:  `ctx := context.TODO()  // Placeholder - fix later`,
			Good: `func DoWork(ctx context.Context) error { ... }  // Accept context`,
			Explanation: "Accept context as the first parameter to allow proper cancellation and deadline propagation.",
		},
		CommonMistakes: []string{
			"WRONG: Replacing TODO() with Background() without considering if cancellation is needed",
			"WRONG: Creating a new context instead of accepting one from the caller",
		},
	},
}

// analyzer implements the contextlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new contextlint analyzer.
func New() ports.Analyzer {
	ctxAnalyzer := &analyzer{}
	ctxAnalyzer.analysis = &analysis.Analyzer{
		Name:     "contextlint",
		Doc:      "detects context misuse common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      ctxAnalyzer.run,
	}
	return ctxAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "contextlint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		call := node.(*ast.CallExpr)

		// Check for context.TODO() call
		if !isContextTODOCall(call) {
			return
		}

		// Skip if in test file
		filename := pass.Fset.Position(call.Pos()).Filename
		if strings.HasSuffix(filename, "_test.go") {
			return
		}

		diag := Diagnostics["AIL010"]
		pass.Report(analysis.Diagnostic{
			Pos:      call.Pos(),
			Category: string(diag.Category),
			Message:  "AIL010: " + diag.Message,
		})
	})

	return nil, nil
}

// isContextTODOCall checks if a call expression is context.TODO().
func isContextTODOCall(call *ast.CallExpr) bool {
	selExpr, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	// Check for context.TODO()
	ident, isIdent := selExpr.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	return ident.Name == "context" && selExpr.Sel.Name == "TODO"
}
