// Package naminglint provides analyzers for detecting Go naming convention issues.
package naminglint

import (
	"go/ast"
	"strings"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
	"github.com/curtbushko/go-ai-lint/internal/core/nolint"
	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Diagnostics contains the diagnostic templates for naminglint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL050": {
		ID:       "AIL050",
		Name:     "getter-with-get-prefix",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryNaming,
		Message:  "getter should not have Get prefix",
		Why: `Go convention is that getters should not have a "Get" prefix.
A method that returns a field "name" should be called Name(), not GetName().
This follows the principle of simplicity and is consistent with the standard library.`,
		Fix: `Remove the "Get" prefix from the method name. Rename GetName() to Name().`,
		Example: domain.FixExample{
			Bad:         `func (u *User) GetName() string { return u.name }`,
			Good:        `func (u *User) Name() string { return u.name }`,
			Explanation: "Go getters use the field name directly without a prefix.",
		},
		CommonMistakes: []string{
			"WRONG: Using GetX() for simple field access",
			"WRONG: Assuming Go follows Java/C# conventions",
		},
	},
	"AIL051": {
		ID:       "AIL051",
		Name:     "redundant-package-prefix",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryNaming,
		Message:  "type name repeats package name",
		Why: `When a type name includes the package name, it creates redundancy.
Users will write user.UserService which stutters. The package name
already provides context, so user.Service is cleaner.`,
		Fix: `Remove the package name prefix from the type name.
Rename UserService to Service when in package "user".`,
		Example: domain.FixExample{
			Bad:         `package user; type UserService struct{}`,
			Good:        `package user; type Service struct{}`,
			Explanation: "Users call user.Service, not user.UserService.",
		},
		CommonMistakes: []string{
			"WRONG: Prefixing all types with package name for 'clarity'",
			"WRONG: Using HttpHTTPServer style double-stuttering",
		},
	},
}

// analyzer implements the naminglint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new naminglint analyzer.
func New() ports.Analyzer {
	namingAnalyzer := &analyzer{}
	namingAnalyzer.analysis = &analysis.Analyzer{
		Name:     "naminglint",
		Doc:      "detects Go naming convention issues common in AI-generated code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      namingAnalyzer.run,
	}
	return namingAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "naminglint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	pkgName := pass.Pkg.Name()

	// Check for AIL050: getter with Get prefix
	funcFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(funcFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)
		a.checkGetterPrefix(pass, funcDecl)
	})

	// Check for AIL051: redundant package prefix in type names
	typeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	insp.Preorder(typeFilter, func(node ast.Node) {
		typeSpec := node.(*ast.TypeSpec)
		a.checkRedundantPackagePrefix(pass, typeSpec, pkgName)
	})

	return nil, nil
}

// checkGetterPrefix checks for AIL050: getter methods with Get prefix.
func (a *analyzer) checkGetterPrefix(pass *analysis.Pass, funcDecl *ast.FuncDecl) {
	// Must be a method (has receiver)
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return
	}

	name := funcDecl.Name.Name

	// Must start with "Get" and have more characters after
	if !strings.HasPrefix(name, "Get") || len(name) <= 3 {
		return
	}

	// Skip common exceptions that are not simple getters
	exceptions := []string{
		"GetOrCreate", "GetOrDefault", "GetOrSet",
		"GetContext", "GetByID", "GetBy", "GetAll",
	}
	for _, exc := range exceptions {
		if strings.HasPrefix(name, exc) {
			return
		}
	}

	// Check if it's a simple getter: no parameters (except receiver)
	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
		return
	}

	// Must return something
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
		return
	}

	// This looks like a simple getter with Get prefix
	diag := Diagnostics["AIL050"]
	nolint.Report(pass, analysis.Diagnostic{
		Pos:      funcDecl.Name.Pos(),
		Category: string(diag.Category),
		Message:  "AIL050: " + diag.Message,
	})
}

// checkRedundantPackagePrefix checks for AIL051: type names that repeat the package name.
func (a *analyzer) checkRedundantPackagePrefix(pass *analysis.Pass, typeSpec *ast.TypeSpec, pkgName string) {
	typeName := typeSpec.Name.Name

	// Only check exported types
	if !ast.IsExported(typeName) {
		return
	}

	// Check if type name starts with package name (case-insensitive)
	pkgNameLower := strings.ToLower(pkgName)
	typeNameLower := strings.ToLower(typeName)

	if strings.HasPrefix(typeNameLower, pkgNameLower) && len(typeName) > len(pkgName) {
		diag := Diagnostics["AIL051"]
		nolint.Report(pass, analysis.Diagnostic{
			Pos:      typeSpec.Name.Pos(),
			Category: string(diag.Category),
			Message:  "AIL051: " + diag.Message,
		})
	}
}
