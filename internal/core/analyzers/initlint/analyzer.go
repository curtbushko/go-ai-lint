// Package initlint provides analyzers for detecting init function issues.
package initlint

import (
	"go/ast"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for initlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL100": {
		ID:       "AIL100",
		Name:     "init-with-network",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryInit,
		Message:  "network call in init",
		Why: `Network calls in init() make package import slow and unreliable.
If the network is unavailable, the entire program fails to start.
init() should be deterministic and fast.`,
		Fix: `Move network initialization to an explicit Init() function that
returns an error, or use lazy initialization with sync.Once.`,
		Example: domain.FixExample{
			Bad: `func init() {
    resp, _ := http.Get("http://api.example.com/config")
    // Parse response...
}`,
			Good: `func Init() error {
    resp, err := http.Get("http://api.example.com/config")
    if err != nil {
        return err
    }
    // Parse response...
    return nil
}`,
			Explanation: "Explicit initialization allows error handling and testing.",
		},
		CommonMistakes: []string{
			"WRONG: Fetching configuration from network in init",
			"WRONG: Connecting to databases in init",
		},
	},
	"AIL101": {
		ID:       "AIL101",
		Name:     "init-with-file-io",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryInit,
		Message:  "file I/O in init",
		Why: `File I/O in init() can fail if files are missing or permissions are
wrong. This makes the package unusable even if the caller doesn't need
that file. It also makes testing harder.`,
		Fix: `Move file operations to an explicit Init() function, or use lazy
initialization with sync.Once.`,
		Example: domain.FixExample{
			Bad: `func init() {
    data, _ := os.ReadFile("config.json")
    // Parse config...
}`,
			Good: `func LoadConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Config{}, err
    }
    // Parse config...
    return cfg, nil
}`,
			Explanation: "Explicit loading allows custom paths and proper error handling.",
		},
		CommonMistakes: []string{
			"WRONG: Reading config files in init",
			"WRONG: Loading certificates or keys in init",
		},
	},
}

// analyzer implements the initlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new initlint analyzer.
func New() ports.Analyzer {
	initAnalyzer := &analyzer{}
	initAnalyzer.analysis = &analysis.Analyzer{
		Name:     "initlint",
		Doc:      "detects init function issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      initAnalyzer.run,
	}
	return initAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "initlint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Find init functions
	funcFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(funcFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		// Only check init functions
		if funcDecl.Name.Name != "init" {
			return
		}

		// Check for network and file I/O calls
		a.checkInitBody(pass, funcDecl.Body)
	})

	return nil, nil
}

// checkInitBody checks the init function body for problematic calls.
func (a *analyzer) checkInitBody(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		callExpr, isCall := node.(*ast.CallExpr)
		if !isCall {
			return true
		}

		// Check for http.* calls (network)
		if a.isHTTPCall(callExpr) {
			diag := Diagnostics["AIL100"]
			pass.Report(analysis.Diagnostic{
				Pos:      callExpr.Pos(),
				Category: string(diag.Category),
				Message:  "AIL100: " + diag.Message,
			})
			return true
		}

		// Check for os.* file I/O calls
		if a.isFileIOCall(callExpr) {
			diag := Diagnostics["AIL101"]
			pass.Report(analysis.Diagnostic{
				Pos:      callExpr.Pos(),
				Category: string(diag.Category),
				Message:  "AIL101: " + diag.Message,
			})
			return true
		}

		return true
	})
}

// isHTTPCall checks if a call is an http package call.
func (a *analyzer) isHTTPCall(call *ast.CallExpr) bool {
	selExpr, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return false
	}

	ident, isIdent := selExpr.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	if ident.Name != "http" {
		return false
	}

	// Check for common HTTP methods
	httpMethods := map[string]bool{
		"Get": true, "Post": true, "Head": true,
		"PostForm": true, "NewRequest": true,
	}

	return httpMethods[selExpr.Sel.Name]
}

// isFileIOCall checks if a call is an os package file I/O call.
func (a *analyzer) isFileIOCall(call *ast.CallExpr) bool {
	selExpr, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return false
	}

	ident, isIdent := selExpr.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	if ident.Name != "os" {
		return false
	}

	// Check for common file I/O methods
	fileIOMethods := map[string]bool{
		"Open": true, "OpenFile": true, "Create": true,
		"ReadFile": true, "WriteFile": true,
		"ReadDir": true, "Mkdir": true, "MkdirAll": true,
		"Remove": true, "RemoveAll": true, "Rename": true,
	}

	return fileIOMethods[selExpr.Sel.Name]
}
