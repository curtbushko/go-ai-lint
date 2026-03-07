// Package stringlint provides analyzers for detecting string handling issues.
package stringlint

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/nolint"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for stringlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL070": {
		ID:       "AIL070",
		Name:     "string-byte-iteration",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryString,
		Message:  "byte iteration on string may break UTF-8 characters",
		Why: `Iterating over a string by index (s[i]) accesses bytes, not runes.
Multi-byte UTF-8 characters will be split incorrectly. For example,
"Hello, 世界" has 13 bytes but only 9 characters.`,
		Fix: `Use range to iterate over runes: for _, r := range s { ... }
Or convert to []rune first if you need index access.`,
		Example: domain.FixExample{
			Bad: `for i := 0; i < len(s); i++ {
    _ = s[i] // Accesses bytes, not characters
}`,
			Good: `for _, r := range s {
    _ = r // Correctly iterates over runes
}`,
			Explanation: "Range on string yields runes (Unicode code points).",
		},
		CommonMistakes: []string{
			"WRONG: Assuming len(s) equals character count",
			"WRONG: Using s[i] to get the i-th character",
		},
	},
	"AIL071": {
		ID:       "AIL071",
		Name:     "string-concat-in-loop",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryString,
		Message:  "string concatenation in loop has O(n^2) complexity",
		Why: `Strings in Go are immutable. Each += creates a new string and copies
all previous content. In a loop with N iterations, this becomes O(N^2)
operations instead of O(N).`,
		Fix: `Use strings.Builder for efficient string building in loops.
Or use strings.Join() if you're joining a slice.`,
		Example: domain.FixExample{
			Bad: `var result string
for _, s := range items {
    result += s // O(n^2) - copies entire string each time
}`,
			Good: `var b strings.Builder
for _, s := range items {
    b.WriteString(s) // O(n) - appends efficiently
}
result := b.String()`,
			Explanation: "strings.Builder uses a []byte internally and grows efficiently.",
		},
		CommonMistakes: []string{
			"WRONG: Using += for building strings in loops",
			"WRONG: Using fmt.Sprintf in a loop to build strings",
		},
	},
}

// analyzer implements the stringlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new stringlint analyzer.
func New() ports.Analyzer {
	stringAnalyzer := &analyzer{}
	stringAnalyzer.analysis = &analysis.Analyzer{
		Name:     "stringlint",
		Doc:      "detects string handling issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      stringAnalyzer.run,
	}
	return stringAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "stringlint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Check for loop issues
	loopFilter := []ast.Node{
		(*ast.ForStmt)(nil),
		(*ast.RangeStmt)(nil),
	}

	insp.Preorder(loopFilter, func(node ast.Node) {
		switch stmt := node.(type) {
		case *ast.ForStmt:
			a.checkForStmt(pass, stmt)
		case *ast.RangeStmt:
			a.checkRangeStmt(pass, stmt)
		}
	})

	return nil, nil
}

// checkForStmt checks for issues in traditional for loops.
func (a *analyzer) checkForStmt(pass *analysis.Pass, forStmt *ast.ForStmt) {
	// Check for AIL070: byte iteration on string
	a.checkByteIteration(pass, forStmt)

	// Check for AIL071: string concatenation in loop
	a.checkStringConcatInLoop(pass, forStmt.Body, forStmt.Pos())
}

// checkRangeStmt checks for issues in range loops.
func (a *analyzer) checkRangeStmt(pass *analysis.Pass, rangeStmt *ast.RangeStmt) {
	// Check for AIL071: string concatenation in loop
	a.checkStringConcatInLoop(pass, rangeStmt.Body, rangeStmt.Pos())
}

// checkByteIteration checks for AIL070: byte iteration pattern on strings.
func (a *analyzer) checkByteIteration(pass *analysis.Pass, forStmt *ast.ForStmt) {
	// Pattern: for i := 0; i < len(s); i++ { ... s[i] ... }
	if forStmt.Init == nil || forStmt.Cond == nil || forStmt.Post == nil {
		return
	}

	// Check condition is i < len(s) or len(s) > i
	stringVar := a.getStringVarFromLenCondition(pass, forStmt.Cond)
	if stringVar == "" {
		return
	}

	// Check if body accesses s[i] where s is the string
	if a.bodyAccessesStringIndex(forStmt.Body, stringVar) {
		diag := Diagnostics["AIL070"]
		nolint.Report(pass, analysis.Diagnostic{
			Pos:      forStmt.Pos(),
			Category: string(diag.Category),
			Message:  "AIL070: " + diag.Message,
		})
	}
}

// getStringVarFromLenCondition extracts the string variable from i < len(s) condition.
func (a *analyzer) getStringVarFromLenCondition(pass *analysis.Pass, cond ast.Expr) string {
	binExpr, isBin := cond.(*ast.BinaryExpr)
	if !isBin {
		return ""
	}

	// Check for < or <= operator
	if binExpr.Op != token.LSS && binExpr.Op != token.LEQ {
		return ""
	}

	// Check if right side is len(s)
	callExpr, isCall := binExpr.Y.(*ast.CallExpr)
	if !isCall {
		return ""
	}

	ident, isIdent := callExpr.Fun.(*ast.Ident)
	if !isIdent || ident.Name != "len" {
		return ""
	}

	if len(callExpr.Args) != 1 {
		return ""
	}

	argIdent, isArgIdent := callExpr.Args[0].(*ast.Ident)
	if !isArgIdent {
		return ""
	}

	// Check if the argument is a string type
	argType := pass.TypesInfo.TypeOf(callExpr.Args[0])
	if argType == nil {
		return ""
	}

	basic, isBasic := argType.Underlying().(*types.Basic)
	if !isBasic || basic.Kind() != types.String {
		return ""
	}

	return argIdent.Name
}

// bodyAccessesStringIndex checks if the loop body accesses s[i].
func (a *analyzer) bodyAccessesStringIndex(body *ast.BlockStmt, stringVar string) bool {
	hasAccess := false

	ast.Inspect(body, func(node ast.Node) bool {
		indexExpr, isIndex := node.(*ast.IndexExpr)
		if !isIndex {
			return true
		}

		ident, isIdent := indexExpr.X.(*ast.Ident)
		if isIdent && ident.Name == stringVar {
			hasAccess = true
			return false
		}

		return true
	})

	return hasAccess
}

// checkStringConcatInLoop checks for AIL071: string += in loop.
func (a *analyzer) checkStringConcatInLoop(pass *analysis.Pass, body *ast.BlockStmt, loopPos token.Pos) {
	hasStringConcat := false

	ast.Inspect(body, func(node ast.Node) bool {
		assignStmt, isAssign := node.(*ast.AssignStmt)
		if !isAssign {
			return true
		}

		// Check for += operator
		if assignStmt.Tok != token.ADD_ASSIGN {
			return true
		}

		// Check if LHS is a string variable
		if len(assignStmt.Lhs) != 1 {
			return true
		}

		lhsType := pass.TypesInfo.TypeOf(assignStmt.Lhs[0])
		if lhsType == nil {
			return true
		}

		basic, isBasic := lhsType.Underlying().(*types.Basic)
		if isBasic && basic.Kind() == types.String {
			hasStringConcat = true
			return false
		}

		return true
	})

	if hasStringConcat {
		diag := Diagnostics["AIL071"]
		nolint.Report(pass, analysis.Diagnostic{
			Pos:      loopPos,
			Category: string(diag.Category),
			Message:  "AIL071: " + diag.Message,
		})
	}
}
