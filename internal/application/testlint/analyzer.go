// Package testlint provides analyzers for detecting test-related issues.
package testlint

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
)

// Diagnostics contains the diagnostic templates for testlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL130": {
		ID:       "AIL130",
		Name:     "missing-testify",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryTest,
		Message:  "test file should use testify for assertions",
		Why: `The testify library provides richer assertion methods and better error
messages than the standard testing package. Using testify leads to more
readable tests and clearer failure messages.

Tests without testify often use manual if-checks with t.Fatal/t.Error,
which are verbose and less expressive.`,
		Fix: `Add testify imports and use its assertion methods:
- require.Equal(t, expected, actual) for fatal assertions
- assert.Equal(t, expected, actual) for non-fatal assertions`,
		Example: domain.FixExample{
			Bad: `func TestSomething(t *testing.T) {
    result := DoSomething()
    if result != expected {
        t.Fatalf("got %v, want %v", result, expected)
    }
}`,
			Good: `import "github.com/stretchr/testify/require"

func TestSomething(t *testing.T) {
    result := DoSomething()
    require.Equal(t, expected, result)
}`,
			Explanation: "Testify provides cleaner assertions with automatic error messages.",
		},
		CommonMistakes: []string{
			"WRONG: Using testify/assert when the test should fail immediately (use require)",
			"WRONG: Mixing manual if-checks with testify in the same test",
		},
	},
	"AIL131": {
		ID:       "AIL131",
		Name:     "raw-t-fail",
		Severity: domain.SeverityLow,
		Category: domain.CategoryTest,
		Message:  "prefer testify assertions over raw t.Errorf/t.Fatalf/t.Error/t.Fatal",
		Why: `When a test file already imports testify, using raw t.Errorf, t.Fatalf,
t.Error, or t.Fatal is inconsistent and misses the benefits of testify's
richer assertions and better error messages. Mixing styles makes tests
harder to read and maintain.`,
		Fix: `Replace raw t.* calls with testify assertions:
- t.Errorf("got %v, want %v", got, want) -> assert.Equal(t, want, got)
- t.Fatalf("got %v, want %v", got, want) -> require.Equal(t, want, got)
- t.Error("message") -> assert.Fail(t, "message")
- t.Fatal("message") -> require.Fail(t, "message")`,
		Example: domain.FixExample{
			Bad: `func TestSomething(t *testing.T) {
    result := DoSomething()
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}`,
			Good: `func TestSomething(t *testing.T) {
    result := DoSomething()
    assert.Equal(t, expected, result)
}`,
			Explanation: "Testify assertions provide automatic diff output and cleaner failure messages.",
		},
		CommonMistakes: []string{
			"WRONG: Using assert.* when the test should fail immediately (use require.*)",
			"WRONG: Converting t.Fatal to assert.Fail (should be require.Fail)",
		},
	},
	"AIL132": {
		ID:       "AIL132",
		Name:     "missing-require-for-setup",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryTest,
		Message:  "use require.* instead of assert.* for setup errors that would invalidate the rest of the test",
		Why: `When checking errors or nil values during test setup, using assert.*
allows the test to continue even when the assertion fails. This can lead to
confusing nil pointer panics or misleading failures later in the test.

Using require.* for setup checks fails the test immediately, providing
clearer error messages and preventing cascading failures.`,
		Fix: `Replace setup assertions with require equivalents:
- assert.NoError(t, err) -> require.NoError(t, err)
- assert.NotNil(t, obj) -> require.NotNil(t, obj)
- assert.Nil(t, err) -> require.Nil(t, err)`,
		Example: domain.FixExample{
			Bad: `func TestSomething(t *testing.T) {
    obj, err := createObject()
    assert.NoError(t, err)
    assert.NotNil(t, obj)
    // If err != nil or obj == nil, this will panic or give misleading errors
    obj.DoSomething()
}`,
			Good: `func TestSomething(t *testing.T) {
    obj, err := createObject()
    require.NoError(t, err)
    require.NotNil(t, obj)
    // Test stops here if setup fails
    obj.DoSomething()
}`,
			Explanation: "require.* stops the test immediately on failure, preventing confusing cascading errors.",
		},
		CommonMistakes: []string{
			"WRONG: Using assert.NoError for setup errors that must succeed for the test to be valid",
			"WRONG: Using assert.NotNil on objects that will be dereferenced later in the test",
		},
	},
}

// analyzer implements the testlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new testlint analyzer.
func New() ports.Analyzer {
	testAnalyzer := &analyzer{}
	testAnalyzer.analysis = &analysis.Analyzer{
		Name:     "testlint",
		Doc:      "detects test-related issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      testAnalyzer.run,
	}
	return testAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "testlint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	// Only analyze test files
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if !strings.HasSuffix(filename, "_test.go") {
			continue
		}

		// Check if this is a k8s e2e test (exception case)
		if isK8sE2ETest(file) {
			continue
		}

		// Check if the file has test functions
		if !hasTestFunctions(file) {
			continue
		}

		// Check if the file imports testify
		if hasTestifyImport(file) {
			// File imports testify - check for AIL131 (raw t.* calls)
			a.checkRawTFail(pass, file)
			// Check for AIL132 (assert.* in setup positions)
			a.checkMissingRequire(pass, file)
		} else {
			// File doesn't import testify - report AIL130
			a.reportMissingTestify(pass, file)
		}
	}

	return nil, nil
}

// isK8sE2ETest checks if the file is a Kubernetes e2e test.
// This includes both the k8s.io/kubernetes e2e framework and the sigs.k8s.io/e2e-framework.
func isK8sE2ETest(file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Path != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			// Check for k8s.io/kubernetes e2e framework
			if strings.Contains(path, "k8s.io/kubernetes/test/e2e/framework") {
				return true
			}
			// Check for sigs.k8s.io/e2e-framework
			if strings.Contains(path, "sigs.k8s.io/e2e-framework") {
				return true
			}
		}
	}
	return false
}

// hasTestifyImport checks if the file imports testify.
func hasTestifyImport(file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Path != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(path, "testify") {
				return true
			}
		}
	}
	return false
}

// hasTestFunctions checks if the file has test functions.
func hasTestFunctions(file *ast.File) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if strings.HasPrefix(fn.Name.Name, "Test") {
			return true
		}
	}
	return false
}

// rawTFailMethods are the testing.T methods that should use testify instead.
var rawTFailMethods = map[string]bool{
	"Errorf": true,
	"Fatalf": true,
	"Error":  true,
	"Fatal":  true,
}

// checkRawTFail checks for raw t.Errorf/t.Fatalf/t.Error/t.Fatal calls
// in test files that import testify.
func (a *analyzer) checkRawTFail(pass *analysis.Pass, file *ast.File) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		// Check if the call is in this file
		callFile := pass.Fset.File(call.Pos())
		if callFile.Name() != pass.Fset.File(file.Pos()).Name() {
			return
		}

		// Check if this is a method call (selector expression)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// Check if the method is one of our target methods
		methodName := sel.Sel.Name
		if !rawTFailMethods[methodName] {
			return
		}

		// Check if the receiver is a testing.T (identifier named t, or any *testing.T)
		if !isTestingTReceiver(pass, sel.X) {
			return
		}

		diag := Diagnostics["AIL131"]
		domain.Report(pass, analysis.Diagnostic{
			Pos:      call.Pos(),
			Category: string(diag.Category),
			Message:  "AIL131: " + diag.Message,
		})
	})
}

// isTestingTReceiver checks if the expression is likely a *testing.T receiver.
// This checks for common patterns: t.Method(), tt.Method(), etc.
func isTestingTReceiver(pass *analysis.Pass, expr ast.Expr) bool {
	// Get the type of the expression
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}

	// Check if the type is *testing.T
	typeStr := tv.Type.String()
	return strings.Contains(typeStr, "testing.T") || strings.Contains(typeStr, "testing.B")
}

// setupAssertMethods maps assert methods that should use require in setup contexts
// to their corresponding diagnostic message.
var setupAssertMethods = map[string]string{
	"NoError": "use require.NoError instead of assert.NoError for setup errors",
	"NotNil":  "use require.NotNil instead of assert.NotNil for setup checks",
	"Nil":     "use require.Nil instead of assert.Nil for setup errors",
}

// checkMissingRequire checks for assert.NoError/NotNil/Nil calls that should
// use require instead because they are checking setup conditions.
func (a *analyzer) checkMissingRequire(pass *analysis.Pass, file *ast.File) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		// Check if the call is in this file
		callFile := pass.Fset.File(call.Pos())
		if callFile.Name() != pass.Fset.File(file.Pos()).Name() {
			return
		}

		// Check if this is a method call (selector expression)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// Check if the receiver is "assert" package
		receiverIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return
		}
		if receiverIdent.Name != "assert" {
			return
		}

		// Check if the method is one of our target methods
		methodName := sel.Sel.Name
		message, isSetupMethod := setupAssertMethods[methodName]
		if !isSetupMethod {
			return
		}

		diag := Diagnostics["AIL132"]
		domain.Report(pass, analysis.Diagnostic{
			Pos:      call.Pos(),
			Category: string(diag.Category),
			Message:  "AIL132: " + message,
		})
	})
}

// reportMissingTestify reports AIL130 for each test function in the file.
func (a *analyzer) reportMissingTestify(pass *analysis.Pass, file *ast.File) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		fn, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		// Only report for test functions
		if !strings.HasPrefix(fn.Name.Name, "Test") {
			return
		}

		// Check if the function is in this file
		fnFile := pass.Fset.File(fn.Pos())
		if fnFile.Name() != pass.Fset.File(file.Pos()).Name() {
			return
		}

		diag := Diagnostics["AIL130"]
		domain.Report(pass, analysis.Diagnostic{
			Pos:      fn.Pos(),
			Category: string(diag.Category),
			Message:  "AIL130: " + diag.Message,
		})
	})
}
