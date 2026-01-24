// Package ports defines the interfaces (contracts) for go-ai-lint.
// Ports are the boundary between the core domain and external adapters.
package ports

import "golang.org/x/tools/go/analysis"

// Analyzer defines the interface for a lint analyzer.
// Implementations provide specific lint checks (e.g., deferlint, contextlint).
type Analyzer interface {
	// Name returns the analyzer name (e.g., "deferlint").
	Name() string

	// Analyzer returns the go/analysis.Analyzer for this linter.
	Analyzer() *analysis.Analyzer
}
