// Package cmdlint provides analyzers for detecting CLI entry point issues.
package cmdlint

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
)

// Diagnostics contains the diagnostic templates for cmdlint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL120": {
		ID:       "AIL120",
		Name:     "missing-cobra",
		Severity: domain.SeverityHigh,
		Category: domain.CategoryCmd,
		Message:  "cmd/main.go should use cobra for CLI structure",
		Why: `CLI applications in Go should use a structured framework like cobra
for command handling. Raw flag parsing becomes unmaintainable as the
application grows, and cobra provides a consistent pattern for
subcommands, flags, and help text that users expect.

AI-generated code often creates simple main() functions with manual
flag parsing, missing the benefits of proper CLI frameworks.`,
		Fix: `Add spf13/cobra as a dependency and structure your CLI with
cobra.Command. Create a root command in cmd/root.go and call it from main().`,
		Example: domain.FixExample{
			Bad: `func main() {
    flag.Parse()
    // manual argument handling
}`,
			Good: `func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}`,
			Explanation: "Use cobra.Command for proper CLI structure with built-in help, flag handling, and subcommand support.",
		},
		CommonMistakes: []string{
			"WRONG: Using flag package for complex CLI applications",
			"WRONG: Defining all commands in main.go instead of separate files",
			"WRONG: Not using cobra's persistent flags for global options",
		},
	},
	"AIL121": {
		ID:       "AIL121",
		Name:     "missing-viper",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryCmd,
		Message:  "cmd/*.go with flag usage should also use viper for configuration",
		Why: `When using flags in CLI applications, viper provides a unified
configuration system that supports environment variables, config files,
and flag binding. This allows 12-factor app compliance and flexible
deployment configurations.

AI-generated code often uses the flag package directly without viper,
missing the benefits of layered configuration management.`,
		Fix: `Add spf13/viper as a dependency and bind your flags to viper.
Use viper.BindPFlag() to connect cobra/flag definitions to viper keys.`,
		Example: domain.FixExample{
			Bad: `var debug = flag.Bool("debug", false, "enable debug")
func main() {
    flag.Parse()
    if *debug { ... }
}`,
			Good: `func init() {
    rootCmd.PersistentFlags().Bool("debug", false, "enable debug")
    viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}
func main() {
    if viper.GetBool("debug") { ... }
}`,
			Explanation: "Use viper to manage configuration from flags, environment variables, and config files in a unified way.",
		},
		CommonMistakes: []string{
			"WRONG: Using flag.* directly without viper binding",
			"WRONG: Not calling viper.AutomaticEnv() for environment variable support",
			"WRONG: Hardcoding configuration values instead of using viper",
		},
	},
	"AIL122": {
		ID:       "AIL122",
		Name:     "cobra-init-in-main",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryCmd,
		Message:  "cobra.Command should be initialized in root.go, not main.go",
		Why: `The main.go file in a cobra-based CLI should only call Execute() on
the root command. The actual cobra.Command initialization should be in
root.go or a similar file. This separation keeps main.go minimal and
follows the standard cobra project structure.

AI-generated code often initializes cobra.Command directly in main.go,
conflating the entry point with command definition.`,
		Fix: `Move cobra.Command initialization to root.go (or cmd/root.go).
Keep main.go minimal - it should only call rootCmd.Execute().`,
		Example: domain.FixExample{
			Bad: `// main.go
var rootCmd = &cobra.Command{
    Use: "myapp",
}
func main() {
    rootCmd.Execute()
}`,
			Good: `// root.go
var rootCmd = &cobra.Command{
    Use: "myapp",
}
func Execute() error {
    return rootCmd.Execute()
}

// main.go
func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}`,
			Explanation: "Separate command definition (root.go) from entry point (main.go) for cleaner project structure.",
		},
		CommonMistakes: []string{
			"WRONG: Defining cobra.Command in main.go",
			"WRONG: Putting all command setup in the main() function",
			"WRONG: Not following standard cobra project layout",
		},
	},
}

// analyzer implements the cmdlint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new cmdlint analyzer.
func New() ports.Analyzer {
	cmdAnalyzer := &analyzer{}
	cmdAnalyzer.analysis = &analysis.Analyzer{
		Name:     "cmdlint",
		Doc:      "detects CLI entry point issues common in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      cmdAnalyzer.run,
	}
	return cmdAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "cmdlint"
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

	// Track if this file is a cmd/**/*.go and check for various issues
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename

		// Check if file is in cmd/**/*.go path
		if !isCmdFile(filename) {
			continue
		}

		// AIL120: Check if file is cmd/**/main.go without cobra import
		if isCmdMainFile(filename) && !hasCobraImport(file) {
			diag := Diagnostics["AIL120"]
			domain.Report(pass, analysis.Diagnostic{
				Pos:      file.Package,
				Category: string(diag.Category),
				Message:  "AIL120: " + diag.Message,
			})
		}

		// AIL121: Check if file has flag usage but no viper import
		if hasFlagUsage(file, insp) && !hasViperImport(file) {
			diag := Diagnostics["AIL121"]
			domain.Report(pass, analysis.Diagnostic{
				Pos:      file.Package,
				Category: string(diag.Category),
				Message:  "AIL121: " + diag.Message,
			})
		}

		// AIL122: Check if cmd/**/main.go initializes cobra.Command
		if isCmdMainFile(filename) && hasCobraCommandInit(file) {
			diag := Diagnostics["AIL122"]
			domain.Report(pass, analysis.Diagnostic{
				Pos:      file.Package,
				Category: string(diag.Category),
				Message:  "AIL122: " + diag.Message,
			})
		}
	}

	return nil, nil
}

// isCmdMainFile checks if the file path matches cmd/**/main.go pattern.
func isCmdMainFile(filename string) bool {
	// Normalize path separators
	filename = filepath.ToSlash(filename)

	// Check if file is named main.go
	if filepath.Base(filename) != "main.go" {
		return false
	}

	// Check if path contains /cmd/ directory
	return strings.Contains(filename, "/cmd/")
}

// hasCobraImport checks if the file imports cobra.
func hasCobraImport(file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Path != nil {
			// Remove quotes from import path
			importPath := strings.Trim(imp.Path.Value, "\"")
			if importPath == "github.com/spf13/cobra" {
				return true
			}
		}
	}
	return false
}

// isCmdFile checks if the file path is within a cmd/ directory.
func isCmdFile(filename string) bool {
	// Normalize path separators
	filename = filepath.ToSlash(filename)

	// Check if path contains /cmd/ directory
	return strings.Contains(filename, "/cmd/")
}

// hasViperImport checks if the file imports viper.
func hasViperImport(file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Path != nil {
			// Remove quotes from import path
			importPath := strings.Trim(imp.Path.Value, "\"")
			if importPath == "github.com/spf13/viper" {
				return true
			}
		}
	}
	return false
}

// hasFlagUsage checks if the file uses the flag package.
func hasFlagUsage(file *ast.File, _ *inspector.Inspector) bool {
	// First check if flag is imported
	hasFlagImport := false
	for _, imp := range file.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if importPath == "flag" {
				hasFlagImport = true
				break
			}
		}
	}

	if !hasFlagImport {
		return false
	}

	// Check for flag.* usage in the AST
	flagUsed := false
	ast.Inspect(file, func(n ast.Node) bool {
		if flagUsed {
			return false
		}

		switch expr := n.(type) {
		case *ast.SelectorExpr:
			// Check for flag.X calls
			if ident, ok := expr.X.(*ast.Ident); ok && ident.Name == "flag" {
				flagUsed = true
				return false
			}
		}
		return true
	})

	return flagUsed
}

// hasCobraCommandInit checks if the file initializes cobra.Command with &cobra.Command{} or cobra.Command{}.
func hasCobraCommandInit(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		// Look for composite literals like &cobra.Command{} or cobra.Command{}
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if the type is cobra.Command
		if isCobraCommandType(compLit.Type) {
			found = true
			return false
		}

		return true
	})

	return found
}

// isCobraCommandType checks if the expression is cobra.Command.
func isCobraCommandType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if it's cobra.Command
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "cobra" && sel.Sel.Name == "Command"
}
