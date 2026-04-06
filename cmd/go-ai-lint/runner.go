package main

import (
	"fmt"
	"io"
	"sort"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

// RunAnalyzers runs the given analyzers on the specified packages.
// It writes diagnostics to stderr and a summary to stdout.
// Returns 0 if no issues were found, 1 otherwise.
func RunAnalyzers(stdout, stderr io.Writer, analyzers []*analysis.Analyzer, args []string) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 {
		args = []string{"."}
	}

	// Load packages
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedName | packages.NeedImports |
			packages.NeedDeps | packages.NeedModule,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "go-ai-lint: %v\n", err)
		return 1
	}

	// Check for package loading errors
	var loadErrors bool
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			_, _ = fmt.Fprintf(stderr, "%v\n", err)
			loadErrors = true
		}
	})
	if loadErrors {
		return 1
	}

	// Run analyzers and collect diagnostics
	var allDiagnostics []diagnostic
	for _, pkg := range pkgs {
		diags := runAnalyzersOnPackage(analyzers, pkg)
		allDiagnostics = append(allDiagnostics, diags...)
	}

	// Sort diagnostics by position
	sort.Slice(allDiagnostics, func(i, j int) bool {
		return allDiagnostics[i].pos < allDiagnostics[j].pos
	})

	// Print diagnostics
	for _, d := range allDiagnostics {
		_, _ = fmt.Fprintf(stderr, "%s\n", d.message)
	}

	// Print summary
	if len(allDiagnostics) == 0 {
		_, _ = fmt.Fprintf(stdout, "go-ai-lint: no issues found\n")
		return 0
	}

	return 1
}

type diagnostic struct {
	pos     string
	message string
}

// runAnalyzersOnPackage runs all analyzers on a single package, handling dependencies.
func runAnalyzersOnPackage(analyzers []*analysis.Analyzer, pkg *packages.Package) []diagnostic {
	if pkg.TypesInfo == nil || len(pkg.Syntax) == 0 {
		return nil
	}

	// Collect all analyzers including dependencies
	allAnalyzers := collectAnalyzers(analyzers)

	// Run analyzers in dependency order
	results := make(map[*analysis.Analyzer]interface{})
	var allDiags []diagnostic

	for _, a := range allAnalyzers {
		diags := runAnalyzer(a, pkg, results)
		// Only collect diagnostics from the requested analyzers, not dependencies
		if isRequestedAnalyzer(a, analyzers) {
			allDiags = append(allDiags, diags...)
		}
	}

	return allDiags
}

// collectAnalyzers returns all analyzers in dependency order (dependencies first).
func collectAnalyzers(analyzers []*analysis.Analyzer) []*analysis.Analyzer {
	seen := make(map[*analysis.Analyzer]bool)
	var result []*analysis.Analyzer

	var visit func(a *analysis.Analyzer)
	visit = func(a *analysis.Analyzer) {
		if seen[a] {
			return
		}
		seen[a] = true
		for _, req := range a.Requires {
			visit(req)
		}
		result = append(result, a)
	}

	for _, a := range analyzers {
		visit(a)
	}

	return result
}

// isRequestedAnalyzer checks if an analyzer is in the requested list.
func isRequestedAnalyzer(a *analysis.Analyzer, requested []*analysis.Analyzer) bool {
	for _, r := range requested {
		if a == r {
			return true
		}
	}
	return false
}

// runAnalyzer runs a single analyzer on a package with its dependencies satisfied.
func runAnalyzer(a *analysis.Analyzer, pkg *packages.Package, results map[*analysis.Analyzer]interface{}) []diagnostic {
	// Build ResultOf map from dependencies
	resultOf := make(map[*analysis.Analyzer]interface{})
	for _, req := range a.Requires {
		resultOf[req] = results[req]
	}

	var diags []diagnostic
	pass := &analysis.Pass{
		Analyzer:     a,
		Fset:         pkg.Fset,
		Files:        pkg.Syntax,
		OtherFiles:   pkg.OtherFiles,
		IgnoredFiles: pkg.IgnoredFiles,
		Pkg:          pkg.Types,
		TypesInfo:    pkg.TypesInfo,
		TypesSizes:   pkg.TypesSizes,
		ResultOf:     resultOf,
		Report: func(d analysis.Diagnostic) {
			pos := pkg.Fset.Position(d.Pos)
			msg := fmt.Sprintf("%s: %s", pos, d.Message)
			diags = append(diags, diagnostic{
				pos:     pos.String(),
				message: msg,
			})
		},
	}

	result, _ := a.Run(pass)
	results[a] = result

	return diags
}

// RunMain is the main entry point that replaces multichecker.Main.
// It runs analyzers and exits with the appropriate code.
func RunMain(analyzers []*analysis.Analyzer, args []string) int {
	return RunAnalyzers(nil, nil, analyzers, args)
}
