// Package deferlint provides analyzers for detecting defer-related issues.
package deferlint

import (
	"go/ast"
	"go/types"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for deferlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL001": {
		ID:       "AIL001",
		Name:     "defer-in-loop",
		Severity: domain.SeverityCritical,
		Category: domain.CategoryDefer,
		Message:  "defer inside loop delays resource cleanup until function returns",
		Why: `Deferred calls accumulate until the function returns. In a loop
processing N items, all N resources stay open simultaneously, risking
resource exhaustion (file descriptors, memory, database connections).

For example, processing 10,000 files would open 10,000 file handles
before closing any of them.`,
		Fix: `Extract the loop body into a separate function. The defer will
then execute after each iteration when the helper function returns.

Alternative: Use an immediately-invoked function literal (closure),
though this adds slight overhead.`,
		Example: domain.FixExample{
			Bad: `for _, f := range files {
    file, _ := os.Open(f)
    defer file.Close()  // All files stay open!
    process(file)
}`,
			Good: `for _, f := range files {
    if err := processFile(f); err != nil {
        return err
    }
}

func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()  // Closes after each file
    return process(file)
}`,
			Explanation: "By extracting to a helper function, defer runs when that function returns (after each iteration), not when the outer function returns.",
		},
		CommonMistakes: []string{
			"WRONG: Removing defer entirely - resource will never be closed on early return or panic",
			"WRONG: Moving defer outside the loop - only closes the last resource assigned to the variable",
			"WRONG: Manually calling Close() without defer - may skip on early return or panic",
		},
	},
	"AIL002": {
		ID:       "AIL002",
		Name:     "defer-close-error-ignored",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryDefer,
		Message:  "deferred Close() error is ignored",
		Why: `Close() can fail, especially for writable files where data may still be
in buffers. Ignoring the error can lead to silent data loss.

For network connections, Close() errors may indicate the connection was
already broken, which could affect application logic.`,
		Fix: `Wrap the Close() call in a closure and handle the error, or use a
named return value to capture the error.`,
		Example: domain.FixExample{
			Bad:  `defer file.Close()  // Error ignored`,
			Good: `defer func() { _ = file.Close() }()  // Explicitly ignored`,
			Explanation: "If you must ignore the error, do so explicitly with _ = to signal intent.",
		},
		CommonMistakes: []string{
			"WRONG: Assuming Close() never fails",
			"WRONG: Logging the error but not affecting the return value for writes",
		},
	},
	"AIL003": {
		ID:       "AIL003",
		Name:     "defer-flush-error-ignored",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryDefer,
		Message:  "deferred Flush()/Sync() error is ignored",
		Why: `Flush() and Sync() are critical for ensuring data is written to disk.
Ignoring their errors can lead to silent data corruption or loss.`,
		Fix: `Handle the Flush()/Sync() error explicitly, especially for important data.`,
		Example: domain.FixExample{
			Bad:  `defer w.Flush()  // Error ignored - data may not be written`,
			Good: `defer func() { err = w.Flush() }()  // Capture error via named return`,
			Explanation: "For buffered writers, Flush errors are critical and should be handled.",
		},
		CommonMistakes: []string{
			"WRONG: Assuming Flush/Sync always succeeds",
			"WRONG: Calling Flush but ignoring its return value",
		},
	},
}

// analyzer implements the deferlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new deferlint analyzer.
func New() ports.Analyzer {
	deferAnalyzer := &analyzer{}
	deferAnalyzer.analysis = &analysis.Analyzer{
		Name:     "deferlint",
		Doc:      "detects defer-related issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      deferAnalyzer.run,
	}
	return deferAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "deferlint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Track loop depth
	loopDepth := 0

	// Node types we care about
	nodeFilter := []ast.Node{
		(*ast.ForStmt)(nil),
		(*ast.RangeStmt)(nil),
		(*ast.DeferStmt)(nil),
		(*ast.FuncLit)(nil),
	}

	// Track function literal depth to handle closures correctly
	funcLitDepth := 0

	insp.Nodes(nodeFilter, func(node ast.Node, push bool) bool {
		switch stmt := node.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			if push {
				loopDepth++
			} else {
				loopDepth--
			}

		case *ast.FuncLit:
			// Function literals reset the loop context
			// because defer in a closure executes when the closure returns
			if push {
				funcLitDepth++
				loopDepth = 0 // Reset - defer in closure is OK
			} else {
				funcLitDepth--
			}

		case *ast.DeferStmt:
			if push {
				// Check AIL001: defer in loop
				if loopDepth > 0 {
					diag := Diagnostics["AIL001"]
					pass.Report(analysis.Diagnostic{
						Pos:      node.Pos(),
						Category: string(diag.Category),
						Message:  "AIL001: " + diag.Message,
					})
				}

				// Check AIL002/AIL003: ignored error from Close/Flush/Sync
				a.checkDeferredErrorIgnored(pass, stmt)
			}
		}
		return true
	})

	// Suppress unused variable warning
	_ = funcLitDepth

	return nil, nil
}

// checkDeferredErrorIgnored checks if a deferred call ignores an error return.
func (a *analyzer) checkDeferredErrorIgnored(pass *analysis.Pass, deferStmt *ast.DeferStmt) {
	// deferStmt.Call is the call expression
	callExpr := deferStmt.Call

	// Check if it's a direct method call (not a closure)
	// If it's a closure, it can handle errors internally
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	methodName := selExpr.Sel.Name

	// Check for Close, Flush, Sync methods
	var diagID string
	switch methodName {
	case "Close":
		diagID = "AIL002"
	case "Flush", "Sync":
		diagID = "AIL003"
	default:
		return
	}

	// Check if the method returns an error
	if !methodReturnsError(pass, callExpr) {
		return
	}

	diag := Diagnostics[diagID]
	pass.Report(analysis.Diagnostic{
		Pos:      deferStmt.Pos(),
		Category: string(diag.Category),
		Message:  diagID + ": " + diag.Message,
	})
}

// methodReturnsError checks if a call expression's function returns an error.
func methodReturnsError(pass *analysis.Pass, call *ast.CallExpr) bool {
	// Get the type of the function being called
	funcType := pass.TypesInfo.TypeOf(call.Fun)
	if funcType == nil {
		return false
	}

	// Get the signature
	sig, ok := funcType.Underlying().(*types.Signature)
	if !ok {
		return false
	}

	// Check if any return value is an error
	results := sig.Results()
	for i := range results.Len() {
		result := results.At(i)
		if isErrorType(result.Type()) {
			return true
		}
	}

	return false
}

// isErrorType checks if a type is the error interface.
func isErrorType(typeToCheck types.Type) bool {
	// Check if type is the error interface
	if named, ok := typeToCheck.(*types.Named); ok {
		return named.Obj().Name() == "error" && named.Obj().Pkg() == nil
	}

	// Also check for interface types that are error
	if iface, ok := typeToCheck.Underlying().(*types.Interface); ok {
		return iface.NumMethods() == 1 && iface.Method(0).Name() == "Error"
	}

	return false
}
