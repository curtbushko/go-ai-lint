// Package optionlint provides analyzers for detecting functional options pattern issues.
package optionlint

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

// Diagnostics contains the diagnostic templates for optionlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL110": {
		ID:       "AIL110",
		Name:     "with-not-option",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryOption,
		Message:  "With* function should return Option type",
		Why: `Functions named With* conventionally follow the functional options pattern.
They should return a function type (Option) that modifies configuration.
Returning primitive types breaks this convention and confuses API users.`,
		Fix: `Define an Option type as func(*Config) and have With* functions
return this type with a closure that applies the configuration.`,
		Example: domain.FixExample{
			Bad: `func WithTimeout(t int) int { return t }`,
			Good: `type Option func(*Config)

func WithTimeout(t int) Option {
    return func(c *Config) { c.Timeout = t }
}`,
			Explanation: "Options are functions that modify config, enabling: New(WithTimeout(5))",
		},
		CommonMistakes: []string{
			"WRONG: With* returning the raw value",
			"WRONG: With* modifying global state directly",
		},
	},
}

// analyzer implements the optionlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new optionlint analyzer.
func New() ports.Analyzer {
	optionAnalyzer := &analyzer{}
	optionAnalyzer.analysis = &analysis.Analyzer{
		Name:     "optionlint",
		Doc:      "detects functional options pattern issues in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      optionAnalyzer.run,
	}
	return optionAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "optionlint"
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

	// Find With* functions
	funcFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(funcFilter, func(node ast.Node) {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}
		a.checkWithFunction(pass, funcDecl)
	})

	return nil, nil
}

// checkWithFunction checks for AIL110: With* not returning Option.
func (a *analyzer) checkWithFunction(pass *analysis.Pass, funcDecl *ast.FuncDecl) {
	name := funcDecl.Name.Name

	// Must start with "With" and have more characters
	if !strings.HasPrefix(name, "With") || len(name) <= 4 {
		return
	}

	// Skip common exceptions that don't follow the pattern (context package)
	exceptions := []string{
		"WithContext", "WithValue", "WithCancel", "WithDeadline",
	}
	for _, exc := range exceptions {
		if name == exc {
			return
		}
	}

	// Must be a function (not a method)
	if funcDecl.Recv != nil {
		return
	}

	// Must have a return type
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
		return
	}

	// Check if return type is a function type (option pattern)
	for _, result := range funcDecl.Type.Results.List {
		resultType := pass.TypesInfo.TypeOf(result.Type)
		if resultType == nil {
			continue
		}

		// Check if it's a function type
		if _, isFunc := resultType.Underlying().(*types.Signature); isFunc {
			return // Good - returns a function
		}

		// Check if it's a named type that is a function
		if named, isNamed := resultType.(*types.Named); isNamed {
			if _, isFunc := named.Underlying().(*types.Signature); isFunc {
				return // Good - returns a named function type (Option)
			}
		}
	}

	// If we get here, the function doesn't return a function type
	diag := Diagnostics["AIL110"]
	domain.Report(pass, analysis.Diagnostic{
		Pos:      funcDecl.Name.Pos(),
		Category: string(diag.Category),
		Message:  "AIL110: " + diag.Message,
	})
}
