// Package errorlint provides analyzers for detecting error handling issues.
package errorlint

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
)

// Diagnostics contains the diagnostic templates for errorlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL030": {
		ID:       "AIL030",
		Name:     "error-handled-twice",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryError,
		Message:  "error handled twice",
		Why: `Logging an error and then returning it causes the error to be handled
twice - once at the current level (logging) and again by the caller.
This leads to duplicate log entries and makes debugging harder.`,
		Fix: `Either log the error OR return it, not both. If you need to add context,
wrap the error with fmt.Errorf and %w, then return it for the caller to log.`,
		Example: domain.FixExample{
			Bad: `if err != nil {
    log.Printf("error: %v", err)
    return err  // Error logged AND returned
}`,
			Good: `if err != nil {
    return fmt.Errorf("operation failed: %w", err)  // Wrap and return only
}`,
			Explanation: "Wrapping adds context without logging. The caller decides whether to log.",
		},
		CommonMistakes: []string{
			"WRONG: Logging at every level creates duplicate log entries",
			"WRONG: Logging without returning loses the error for the caller",
		},
	},
	"AIL033": {
		ID:       "AIL033",
		Name:     "error-fmt-not-wrapped",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryError,
		Message:  "error wrapped with %v instead of %w",
		Why: `Using %v instead of %w in fmt.Errorf breaks the error chain.
The original error is converted to a string, making errors.Is() and
errors.As() unable to match the underlying error.`,
		Fix: `Use %w instead of %v when wrapping errors with fmt.Errorf.
This preserves the error chain for proper error handling.`,
		Example: domain.FixExample{
			Bad:         `return fmt.Errorf("failed: %v", err)  // Breaks error chain`,
			Good:        `return fmt.Errorf("failed: %w", err)  // Preserves error chain`,
			Explanation: "With %w, errors.Is(returnedErr, originalErr) returns true.",
		},
		CommonMistakes: []string{
			"WRONG: Using %v when you want to preserve the error for Is/As checks",
			"WRONG: Using %s which also breaks the chain",
		},
	},
}

// analyzer implements the errorlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new errorlint analyzer.
func New() ports.Analyzer {
	errorAnalyzer := &analyzer{}
	errorAnalyzer.analysis = &analysis.Analyzer{
		Name:     "errorlint",
		Doc:      "detects error handling issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      errorAnalyzer.run,
	}
	return errorAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "errorlint"
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

	// Check for AIL030: error handled twice
	ifStmtFilter := []ast.Node{
		(*ast.IfStmt)(nil),
	}

	insp.Preorder(ifStmtFilter, func(node ast.Node) {
		ifStmt, ok := node.(*ast.IfStmt)
		if !ok {
			return
		}
		a.checkErrorHandledTwice(pass, ifStmt)
	})

	// Check for AIL033: error wrapped with %v instead of %w
	callFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(callFilter, func(node ast.Node) {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}
		a.checkErrorFmtNotWrapped(pass, callExpr)
	})

	return nil, nil
}

// checkErrorHandledTwice checks for AIL030: error logged and returned.
func (a *analyzer) checkErrorHandledTwice(pass *analysis.Pass, ifStmt *ast.IfStmt) {
	// Check if condition is "err != nil"
	errVar := a.getErrorVarFromCondition(ifStmt.Cond)
	if errVar == "" {
		return
	}

	// Check if body contains both log call and return with error
	hasLogCall := false
	hasReturnError := false
	var loggedErrVar string
	var returnedErrVar string

	ast.Inspect(ifStmt.Body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.ExprStmt:
			// Check for log.Print*, log.Fatal*, etc.
			if callExpr, isCall := stmt.X.(*ast.CallExpr); isCall {
				if a.isLogCall(callExpr) {
					loggedErr := a.getErrorArgFromCall(callExpr)
					if loggedErr != "" {
						hasLogCall = true
						loggedErrVar = loggedErr
					}
				}
			}
		case *ast.ReturnStmt:
			// Check if return includes the error
			for _, result := range stmt.Results {
				if a.containsErrorVar(result, errVar) {
					hasReturnError = true
					returnedErrVar = errVar
				}
			}
		}
		return true
	})

	// Report if both logging and returning the same error
	if hasLogCall && hasReturnError && (loggedErrVar == returnedErrVar || loggedErrVar == errVar) {
		diag := Diagnostics["AIL030"]
		domain.Report(pass, analysis.Diagnostic{
			Pos:      ifStmt.Cond.Pos(),
			Category: string(diag.Category),
			Message:  "AIL030: " + diag.Message,
		})
	}
}

// getErrorVarFromCondition extracts error variable name from "err != nil" condition.
func (a *analyzer) getErrorVarFromCondition(cond ast.Expr) string {
	binExpr, isBin := cond.(*ast.BinaryExpr)
	if !isBin {
		return ""
	}

	// Check for != nil
	if binExpr.Op.String() != "!=" {
		return ""
	}

	// Check if one side is nil
	var errSide ast.Expr
	if ident, isIdent := binExpr.Y.(*ast.Ident); isIdent && ident.Name == "nil" {
		errSide = binExpr.X
	} else if ident, isIdent := binExpr.X.(*ast.Ident); isIdent && ident.Name == "nil" {
		errSide = binExpr.Y
	} else {
		return ""
	}

	// Get the error variable name
	if ident, isIdent := errSide.(*ast.Ident); isIdent {
		return ident.Name
	}

	return ""
}

// isLogCall checks if a call is a log function (log.Print, log.Printf, etc.).
func (a *analyzer) isLogCall(call *ast.CallExpr) bool {
	selExpr, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return false
	}

	ident, isIdent := selExpr.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	if ident.Name != "log" {
		return false
	}

	// Check for common log functions
	logFuncs := map[string]bool{
		"Print": true, "Printf": true, "Println": true,
		"Fatal": true, "Fatalf": true, "Fatalln": true,
		"Panic": true, "Panicf": true, "Panicln": true,
	}

	return logFuncs[selExpr.Sel.Name]
}

// getErrorArgFromCall extracts error variable name from log call arguments.
func (a *analyzer) getErrorArgFromCall(call *ast.CallExpr) string {
	for _, arg := range call.Args {
		if ident, isIdent := arg.(*ast.Ident); isIdent {
			// Assume variable names containing "err" are errors
			if strings.Contains(strings.ToLower(ident.Name), "err") {
				return ident.Name
			}
		}
	}
	return ""
}

// containsErrorVar checks if an expression contains the error variable.
func (a *analyzer) containsErrorVar(expr ast.Expr, errVar string) bool {
	// Direct identifier
	if ident, isIdent := expr.(*ast.Ident); isIdent {
		return ident.Name == errVar
	}

	// Check in fmt.Errorf wrapping
	if callExpr, isCall := expr.(*ast.CallExpr); isCall {
		for _, arg := range callExpr.Args {
			if ident, isIdent := arg.(*ast.Ident); isIdent && ident.Name == errVar {
				return true
			}
		}
	}

	return false
}

// checkErrorFmtNotWrapped checks for AIL033: fmt.Errorf with %v instead of %w.
func (a *analyzer) checkErrorFmtNotWrapped(pass *analysis.Pass, call *ast.CallExpr) {
	// Check if this is fmt.Errorf
	if !a.isFmtErrorf(call) {
		return
	}

	// Need at least 2 args: format string and error
	if len(call.Args) < 2 {
		return
	}

	// Get format string
	formatLit, isLit := call.Args[0].(*ast.BasicLit)
	if !isLit {
		return
	}
	formatStr := formatLit.Value

	// Find error arguments (not the format string)
	for argIdx := 1; argIdx < len(call.Args); argIdx++ {
		arg := call.Args[argIdx]

		// Check if this argument is an error type
		if !a.isErrorType(pass, arg) {
			continue
		}

		// Check if the corresponding format verb is %v or %s instead of %w
		if a.hasNonWrappingVerbForError(formatStr, argIdx) {
			diag := Diagnostics["AIL033"]
			domain.Report(pass, analysis.Diagnostic{
				Pos:      call.Pos(),
				Category: string(diag.Category),
				Message:  "AIL033: " + diag.Message,
			})
			return // Only report once per call
		}
	}
}

// isFmtErrorf checks if a call is fmt.Errorf.
func (a *analyzer) isFmtErrorf(call *ast.CallExpr) bool {
	selExpr, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return false
	}

	ident, isIdent := selExpr.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	return ident.Name == "fmt" && selExpr.Sel.Name == "Errorf"
}

// isErrorType checks if an expression is of error type.
func (a *analyzer) isErrorType(pass *analysis.Pass, expr ast.Expr) bool {
	exprType := pass.TypesInfo.TypeOf(expr)
	if exprType == nil {
		return false
	}

	// Check if it implements error interface
	errorObj := types.Universe.Lookup("error")
	if errorObj == nil {
		return false
	}
	errorInterface, ok := errorObj.Type().Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(exprType, errorInterface)
}

// hasNonWrappingVerbForError checks if format string uses %v or %s for the error at given position.
func (a *analyzer) hasNonWrappingVerbForError(formatStr string, argPosition int) bool {
	// Simple parsing: count format verbs and check if the one at argPosition is %v or %s
	// argPosition is 1-based (first arg after format string is position 1)
	verbs := a.extractFormatVerbs(formatStr)

	if argPosition > 0 && argPosition <= len(verbs) {
		verb := verbs[argPosition-1]
		return verb == 'v' || verb == 's'
	}

	return false
}

// extractFormatVerbs extracts all format verbs from a format string.
func (a *analyzer) extractFormatVerbs(formatStr string) []byte {
	var verbs []byte

	//nolint:stringlint // AIL070: Intentional byte iteration - format verbs are ASCII only
	for charIdx := 0; charIdx < len(formatStr); charIdx++ {
		if formatStr[charIdx] != '%' {
			continue
		}

		if charIdx+1 >= len(formatStr) {
			break
		}

		// Check for escaped %
		if formatStr[charIdx+1] == '%' {
			charIdx++
			continue
		}

		// Find the verb character (skip flags, width, precision)
		verb, newIdx := a.findVerbChar(formatStr, charIdx+1)
		if verb != 0 {
			verbs = append(verbs, verb)
			charIdx = newIdx
		}
	}

	return verbs
}

// findVerbChar finds the verb character starting from the given index.
func (a *analyzer) findVerbChar(formatStr string, startIdx int) (byte, int) {
	//nolint:stringlint // AIL070: Intentional byte iteration - verbs are ASCII letters only
	for scanIdx := startIdx; scanIdx < len(formatStr); scanIdx++ {
		char := formatStr[scanIdx]
		// Check if this is a verb character (letter)
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			return char, scanIdx
		}
	}
	return 0, startIdx
}
