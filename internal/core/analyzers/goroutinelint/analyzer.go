// Package goroutinelint provides analyzers for detecting goroutine lifecycle issues.
package goroutinelint

import (
	"go/ast"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for goroutinelint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL020": {
		ID:       "AIL020",
		Name:     "goroutine-no-cancel",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryGoroutine,
		Message:  "goroutine without cancellation mechanism",
		Why: `Goroutines that lack a cancellation mechanism will run indefinitely,
leading to goroutine leaks. This is a common source of memory leaks and
resource exhaustion in long-running applications.`,
		Fix: `Add a context parameter and check ctx.Done() in the goroutine's loop,
or use a done channel to signal when the goroutine should exit.`,
		Example: domain.FixExample{
			Bad: `go func() {
    for {
        doWork()
    }
}()`,
			Good: `go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
        }
    }
}()`,
			Explanation: "Use context or done channel to signal goroutine termination.",
		},
		CommonMistakes: []string{
			"WRONG: Ignoring cancellation entirely and relying on program exit",
			"WRONG: Using a boolean flag without proper synchronization",
		},
	},
	"AIL021": {
		ID:       "AIL021",
		Name:     "goroutine-infinite-loop",
		Severity: domain.SeverityCritical,
		Category: domain.CategoryGoroutine,
		Message:  "infinite loop in goroutine without exit condition",
		Why: `An infinite loop without any exit condition (break, return, or channel receive)
will never terminate, guaranteeing a goroutine leak.`,
		Fix: `Add an exit condition: check a done channel, context cancellation,
or add a break/return statement based on a condition.`,
		Example: domain.FixExample{
			Bad: `go func() {
    for {
        doWork()
    }
}()`,
			Good: `go func() {
    for {
        select {
        case <-done:
            return
        default:
            doWork()
        }
    }
}()`,
			Explanation: "Always provide an exit path for infinite loops in goroutines.",
		},
		CommonMistakes: []string{
			"WRONG: Assuming the goroutine will exit when main returns",
			"WRONG: Using break without a condition that can become true",
		},
	},
	"AIL022": {
		ID:       "AIL022",
		Name:     "goroutine-closure-capture",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryGoroutine,
		Message:  "loop variable captured by goroutine",
		Why: `In Go versions before 1.22, loop variables are reused across iterations.
When captured by a goroutine closure, all goroutines may see the same
(final) value instead of the expected per-iteration value.`,
		Fix: `Either shadow the loop variable with a local copy, or pass it
as a parameter to the goroutine function.`,
		Example: domain.FixExample{
			Bad: `for _, item := range items {
    go func() {
        process(item) // All goroutines may see the last item
    }()
}`,
			Good: `for _, item := range items {
    item := item // Shadow the variable
    go func() {
        process(item)
    }()
}`,
			Explanation: "Shadowing creates a new variable per iteration that is safely captured.",
		},
		CommonMistakes: []string{
			"WRONG: Assuming each iteration gets a fresh variable (only true in Go 1.22+)",
			"WRONG: Using a pointer to the loop variable",
		},
	},
}

// analyzer implements the goroutinelint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new goroutinelint analyzer.
func New() ports.Analyzer {
	goroutineAnalyzer := &analyzer{}
	goroutineAnalyzer.analysis = &analysis.Analyzer{
		Name:     "goroutinelint",
		Doc:      "detects goroutine lifecycle issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      goroutineAnalyzer.run,
	}
	return goroutineAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "goroutinelint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// We need to track loop variables for AIL022
	nodeFilter := []ast.Node{
		(*ast.RangeStmt)(nil),
		(*ast.ForStmt)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		switch loopStmt := node.(type) {
		case *ast.RangeStmt:
			a.checkRangeLoopCapture(pass, loopStmt)
		case *ast.ForStmt:
			a.checkForLoopCapture(pass, loopStmt)
		}
	})

	// Check for goroutines with infinite loops
	goStmtFilter := []ast.Node{
		(*ast.GoStmt)(nil),
	}

	insp.Preorder(goStmtFilter, func(node ast.Node) {
		goStmt := node.(*ast.GoStmt)
		a.checkGoroutine(pass, goStmt)
	})

	return nil, nil
}

// checkGoroutine checks a goroutine for AIL020 and AIL021.
func (a *analyzer) checkGoroutine(pass *analysis.Pass, goStmt *ast.GoStmt) {
	// Get the function body
	funcLit, isFuncLit := goStmt.Call.Fun.(*ast.FuncLit)
	if !isFuncLit {
		return
	}

	// Check for infinite loops in the goroutine body
	hasInfiniteLoop := false
	hasExitPath := false     // conditional break/return
	hasCancellation := false // select with channel
	hasTimeSleep := false    // time.Sleep or similar pacing

	ast.Inspect(funcLit.Body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.ForStmt:
			// Check if this is an infinite loop (no condition)
			if stmt.Cond == nil {
				hasInfiniteLoop = true
				// Check if the loop body has exit paths
				hasExitPath = hasExitPath || a.hasConditionalExit(stmt.Body)
				hasCancellation = hasCancellation || a.hasCancellationCheck(stmt.Body)
				hasTimeSleep = hasTimeSleep || a.hasTimeSleep(stmt.Body)
			}
		case *ast.RangeStmt:
			// Range loops are finite, so they're okay
		}
		return true
	})

	// If there's a cancellation mechanism, everything is fine
	if hasCancellation {
		return
	}

	// If there's a conditional exit path (if + break/return), it's acceptable
	if hasExitPath {
		return
	}

	// No exit path at all - check severity
	if hasInfiniteLoop {
		if hasTimeSleep {
			// Has time.Sleep - developer intended long-running but forgot cancellation
			diag := Diagnostics["AIL020"]
			pass.Report(analysis.Diagnostic{
				Pos:      goStmt.Pos(),
				Category: string(diag.Category),
				Message:  "AIL020: " + diag.Message,
			})
		} else {
			// Pure spin loop with no exit at all
			diag := Diagnostics["AIL021"]
			pass.Report(analysis.Diagnostic{
				Pos:      goStmt.Pos(),
				Category: string(diag.Category),
				Message:  "AIL021: " + diag.Message,
			})
		}
	}
}

// hasConditionalExit checks if a block has a conditional exit (if + break/return).
func (a *analyzer) hasConditionalExit(block *ast.BlockStmt) bool {
	hasExit := false
	ast.Inspect(block, func(node ast.Node) bool {
		ifStmt, isIf := node.(*ast.IfStmt)
		if !isIf {
			return true
		}
		// Check if if statement contains break or return
		if a.containsBreakOrReturn(ifStmt.Body) {
			hasExit = true
		}
		return !hasExit
	})
	return hasExit
}

// hasTimeSleep checks if a block contains time.Sleep or similar pacing calls.
func (a *analyzer) hasTimeSleep(block *ast.BlockStmt) bool {
	hasSleep := false
	ast.Inspect(block, func(node ast.Node) bool {
		callExpr, isCall := node.(*ast.CallExpr)
		if !isCall {
			return true
		}

		// Check for time.Sleep
		selExpr, isSel := callExpr.Fun.(*ast.SelectorExpr)
		if !isSel {
			return true
		}

		ident, isIdent := selExpr.X.(*ast.Ident)
		if !isIdent {
			return true
		}

		// Check for time.Sleep, time.After, time.Tick, etc.
		if ident.Name == "time" {
			switch selExpr.Sel.Name {
			case "Sleep", "After", "Tick", "NewTimer", "NewTicker":
				hasSleep = true
			}
		}

		return !hasSleep
	})
	return hasSleep
}

// containsBreakOrReturn checks if a block contains break or return.
func (a *analyzer) containsBreakOrReturn(block *ast.BlockStmt) bool {
	hasBreakOrReturn := false
	ast.Inspect(block, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.BranchStmt:
			if stmt.Tok.String() == "break" || stmt.Tok.String() == "return" {
				hasBreakOrReturn = true
			}
		case *ast.ReturnStmt:
			hasBreakOrReturn = true
		}
		return !hasBreakOrReturn
	})
	return hasBreakOrReturn
}

// hasCancellationCheck checks if a block has cancellation checking (ctx.Done() or done channel).
func (a *analyzer) hasCancellationCheck(block *ast.BlockStmt) bool {
	hasCancellation := false
	ast.Inspect(block, func(node ast.Node) bool {
		selectStmt, isSelect := node.(*ast.SelectStmt)
		if !isSelect {
			return true
		}
		// Check if select has a case receiving from a channel
		for _, clause := range selectStmt.Body.List {
			commClause, isComm := clause.(*ast.CommClause)
			if !isComm || commClause.Comm == nil {
				continue
			}
			// This is a channel receive case, which is a cancellation mechanism
			hasCancellation = true
		}
		return !hasCancellation
	})
	return hasCancellation
}

// checkRangeLoopCapture checks for AIL022 in range loops.
func (a *analyzer) checkRangeLoopCapture(pass *analysis.Pass, rangeStmt *ast.RangeStmt) {
	// Get loop variables
	loopVars := make(map[string]bool)
	if ident, isIdent := rangeStmt.Key.(*ast.Ident); isIdent && ident.Name != "_" {
		loopVars[ident.Name] = true
	}
	if rangeStmt.Value != nil {
		if ident, isIdent := rangeStmt.Value.(*ast.Ident); isIdent && ident.Name != "_" {
			loopVars[ident.Name] = true
		}
	}

	if len(loopVars) == 0 {
		return
	}

	// Find go statements in the loop body
	ast.Inspect(rangeStmt.Body, func(node ast.Node) bool {
		goStmt, isGoStmt := node.(*ast.GoStmt)
		if !isGoStmt {
			return true
		}

		funcLit, isFuncLit := goStmt.Call.Fun.(*ast.FuncLit)
		if !isFuncLit {
			return true
		}

		// Check if any loop variables are captured
		capturedVars := a.findCapturedLoopVars(funcLit, loopVars, rangeStmt.Body)
		for varName := range capturedVars {
			diag := Diagnostics["AIL022"]
			pass.Report(analysis.Diagnostic{
				Pos:      goStmt.Pos(),
				Category: string(diag.Category),
				Message:  "AIL022: loop variable '" + varName + "' captured by goroutine",
			})
		}

		return true
	})
}

// checkForLoopCapture checks for AIL022 in for loops.
func (a *analyzer) checkForLoopCapture(pass *analysis.Pass, forStmt *ast.ForStmt) {
	// Get loop variables from init statement
	loopVars := make(map[string]bool)
	if forStmt.Init != nil {
		if assignStmt, isAssign := forStmt.Init.(*ast.AssignStmt); isAssign {
			for _, lhs := range assignStmt.Lhs {
				if ident, isIdent := lhs.(*ast.Ident); isIdent {
					loopVars[ident.Name] = true
				}
			}
		}
	}

	if len(loopVars) == 0 {
		return
	}

	// Find go statements in the loop body
	ast.Inspect(forStmt.Body, func(node ast.Node) bool {
		goStmt, isGoStmt := node.(*ast.GoStmt)
		if !isGoStmt {
			return true
		}

		funcLit, isFuncLit := goStmt.Call.Fun.(*ast.FuncLit)
		if !isFuncLit {
			return true
		}

		// Check if any loop variables are captured
		capturedVars := a.findCapturedLoopVars(funcLit, loopVars, forStmt.Body)
		for varName := range capturedVars {
			diag := Diagnostics["AIL022"]
			pass.Report(analysis.Diagnostic{
				Pos:      goStmt.Pos(),
				Category: string(diag.Category),
				Message:  "AIL022: loop variable '" + varName + "' captured by goroutine",
			})
		}

		return true
	})
}

// findCapturedLoopVars finds loop variables that are captured by a closure.
func (a *analyzer) findCapturedLoopVars(funcLit *ast.FuncLit, loopVars map[string]bool, loopBody *ast.BlockStmt) map[string]bool {
	capturedVars := make(map[string]bool)

	// Check if variables are passed as parameters to the function
	paramNames := make(map[string]bool)
	for _, param := range funcLit.Type.Params.List {
		for _, name := range param.Names {
			paramNames[name.Name] = true
		}
	}

	// Check if variables are shadowed before the goroutine
	shadowedVars := a.findShadowedVars(loopBody, funcLit)

	// Find identifiers used in the function body
	ast.Inspect(funcLit.Body, func(node ast.Node) bool {
		ident, isIdent := node.(*ast.Ident)
		if !isIdent {
			return true
		}

		// Skip if this is a parameter
		if paramNames[ident.Name] {
			return true
		}

		// Skip if this variable was shadowed
		if shadowedVars[ident.Name] {
			return true
		}

		// Check if this is a loop variable
		if loopVars[ident.Name] {
			capturedVars[ident.Name] = true
		}

		return true
	})

	return capturedVars
}

// findShadowedVars finds variables that are shadowed before a given position.
func (a *analyzer) findShadowedVars(loopBody *ast.BlockStmt, funcLit *ast.FuncLit) map[string]bool {
	shadowedVars := make(map[string]bool)

	// Look for assignments like `item := item` before the funcLit
	for _, stmt := range loopBody.List {
		// Stop when we reach the go statement containing funcLit
		if a.containsFuncLit(stmt, funcLit) {
			break
		}

		assignStmt, isAssign := stmt.(*ast.AssignStmt)
		if !isAssign || assignStmt.Tok.String() != ":=" {
			continue
		}

		// Check if this is a shadowing assignment (x := x)
		for idx, lhs := range assignStmt.Lhs {
			lhsIdent, isLHSIdent := lhs.(*ast.Ident)
			if !isLHSIdent {
				continue
			}
			if idx < len(assignStmt.Rhs) {
				rhsIdent, isRHSIdent := assignStmt.Rhs[idx].(*ast.Ident)
				if isRHSIdent && lhsIdent.Name == rhsIdent.Name {
					shadowedVars[lhsIdent.Name] = true
				}
			}
		}
	}

	return shadowedVars
}

// containsFuncLit checks if a statement contains the given function literal.
func (a *analyzer) containsFuncLit(stmt ast.Stmt, target *ast.FuncLit) bool {
	found := false
	ast.Inspect(stmt, func(node ast.Node) bool {
		if node == target {
			found = true
			return false
		}
		return true
	})
	return found
}
