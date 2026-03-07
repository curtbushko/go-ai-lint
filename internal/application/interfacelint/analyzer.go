// Package interfacelint provides analyzers for detecting interface design issues.
package interfacelint

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
)

const maxInterfaceMethods = 5

// Diagnostics contains the diagnostic templates for interfacelint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL040": {
		ID:       "AIL040",
		Name:     "interface-too-large",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryInterface,
		Message:  "interface has too many methods",
		Why: `Large interfaces (more than 5 methods) violate the Interface Segregation
Principle. They are harder to implement, mock, and test. Go prefers
small, focused interfaces that can be composed.`,
		Fix: `Split the interface into smaller, focused interfaces. Use interface
embedding to compose them when needed.`,
		Example: domain.FixExample{
			Bad: `type Repository interface {
    Create(); Read(); Update(); Delete(); List(); Search(); Count()
}`,
			Good: `type Reader interface { Read() }
type Writer interface { Create(); Update(); Delete() }
type Searcher interface { List(); Search(); Count() }`,
			Explanation: "Small interfaces can be implemented and composed independently.",
		},
		CommonMistakes: []string{
			"WRONG: Creating one interface per service/component",
			"WRONG: Including every method a type might need",
		},
	},
	"AIL042": {
		ID:       "AIL042",
		Name:     "interface-missing-er-suffix",
		Severity: domain.SeverityLow,
		Category: domain.CategoryInterface,
		Message:  "single-method interface should have -er suffix",
		Why: `Go convention for single-method interfaces is to name them with an
"-er" suffix based on the method: Read -> Reader, Write -> Writer,
Close -> Closer. This makes the interface's purpose immediately clear.`,
		Fix: `Rename the interface to use the "-er" suffix. For example,
rename "Validate" to "Validator".`,
		Example: domain.FixExample{
			Bad:         `type Validate interface { Validate() error }`,
			Good:        `type Validator interface { Validate() error }`,
			Explanation: "The -er suffix indicates this interface wraps a single action.",
		},
		CommonMistakes: []string{
			"WRONG: Naming interfaces after the action without -er",
			"WRONG: Using -able suffix instead of -er for single methods",
		},
	},
}

// analyzer implements the interfacelint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new interfacelint analyzer.
func New() ports.Analyzer {
	interfaceAnalyzer := &analyzer{}
	interfaceAnalyzer.analysis = &analysis.Analyzer{
		Name:     "interfacelint",
		Doc:      "detects interface design issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      interfaceAnalyzer.run,
	}
	return interfaceAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "interfacelint"
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

	// Check for interface issues
	typeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	insp.Preorder(typeFilter, func(node ast.Node) {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return
		}
		interfaceType, isInterface := typeSpec.Type.(*ast.InterfaceType)
		if !isInterface {
			return
		}

		a.checkInterfaceSize(pass, typeSpec, interfaceType)
		a.checkSingleMethodNaming(pass, typeSpec, interfaceType)
	})

	return nil, nil
}

// checkInterfaceSize checks for AIL040: interface with too many methods.
func (a *analyzer) checkInterfaceSize(pass *analysis.Pass, typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType) {
	if interfaceType.Methods == nil {
		return
	}

	methodCount := len(interfaceType.Methods.List)
	if methodCount > maxInterfaceMethods {
		diag := Diagnostics["AIL040"]
		domain.Report(pass, analysis.Diagnostic{
			Pos:      typeSpec.Name.Pos(),
			Category: string(diag.Category),
			Message:  fmt.Sprintf("AIL040: interface has %d methods (max %d)", methodCount, maxInterfaceMethods),
		})
	}
}

// checkSingleMethodNaming checks for AIL042: single-method interface without -er suffix.
func (a *analyzer) checkSingleMethodNaming(pass *analysis.Pass, typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType) {
	if interfaceType.Methods == nil {
		return
	}

	// Must have exactly one method
	if len(interfaceType.Methods.List) != 1 {
		return
	}

	interfaceName := typeSpec.Name.Name

	// Check if the interface name already ends with common -er patterns
	erSuffixes := []string{"er", "or", "ar"}
	for _, suffix := range erSuffixes {
		if strings.HasSuffix(strings.ToLower(interfaceName), suffix) {
			return
		}
	}

	diag := Diagnostics["AIL042"]
	domain.Report(pass, analysis.Diagnostic{
		Pos:      typeSpec.Name.Pos(),
		Category: string(diag.Category),
		Message:  "AIL042: single-method interface should have -er suffix",
	})
}
