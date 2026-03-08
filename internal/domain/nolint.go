package domain

// This file provides utilities for parsing and checking nolint directives.
// Supports: //nolint, //nolint:analyzer, //nolint:analyzer1,analyzer2

import (
	"go/ast"
	"go/token"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// nolintConfig holds the global nolint configuration.
// Access is thread-safe via sync.RWMutex.
var (
	nolintEnabled = true
	nolintMu      sync.RWMutex
)

// SetNolintEnabled sets whether nolint directives are enabled globally.
// When false, all nolint directives are ignored and issues are always reported.
// This function is thread-safe.
func SetNolintEnabled(enabled bool) {
	nolintMu.Lock()
	defer nolintMu.Unlock()
	nolintEnabled = enabled
}

// IsNolintEnabled returns whether nolint directives are enabled.
// This function is thread-safe.
func IsNolintEnabled() bool {
	nolintMu.RLock()
	defer nolintMu.RUnlock()
	return nolintEnabled
}

// ParseDirective parses a nolint comment and returns whether it applies to all
// analyzers or a specific list of analyzers.
// Returns (true, nil) for //nolint (all analyzers).
// Returns (false, []string{...}) for //nolint:analyzer1,analyzer2.
// Returns (false, nil) if not a nolint directive.
func ParseDirective(comment string) (all bool, analyzers []string) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return false, nil
	}

	// Remove leading //
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimSpace(comment)

	// Check if it starts with "nolint"
	if !strings.HasPrefix(comment, "nolint") {
		return false, nil
	}

	// Check for //nolint without colon (applies to all)
	rest := strings.TrimPrefix(comment, "nolint")
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return true, nil
	}

	// Check for //nolint:analyzers
	if !strings.HasPrefix(rest, ":") {
		return false, nil
	}

	// Parse analyzer list
	rest = strings.TrimPrefix(rest, ":")

	// Handle trailing comments: //nolint:analyzer // comment
	if idx := strings.Index(rest, "//"); idx != -1 {
		rest = rest[:idx]
	}

	rest = strings.TrimSpace(rest)
	if rest == "" {
		return true, nil
	}

	parts := strings.Split(rest, ",")
	analyzers = make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			analyzers = append(analyzers, p)
		}
	}

	return false, analyzers
}

// ShouldSkip checks if a diagnostic at the given position should be skipped
// due to a nolint directive. It checks comments on the same line and the line above.
// If nolint directives are globally disabled via SetNolintEnabled(false),
// this function always returns false (never skips).
func ShouldSkip(fset *token.FileSet, comments []*ast.CommentGroup, pos token.Pos, analyzerName string) bool {
	// If nolint is globally disabled, never skip any issues
	if !IsNolintEnabled() {
		return false
	}

	position := fset.Position(pos)
	targetLine := position.Line

	for _, cg := range comments {
		for _, c := range cg.List {
			commentPos := fset.Position(c.Pos())
			commentLine := commentPos.Line

			// Check same line or line above
			if commentLine != targetLine && commentLine != targetLine-1 {
				continue
			}

			all, analyzers := ParseDirective(c.Text)
			if all {
				return true
			}

			for _, a := range analyzers {
				if a == analyzerName {
					return true
				}
			}
		}
	}

	return false
}

// Report reports a diagnostic if not suppressed by a nolint directive.
// Use this instead of pass.Report() to respect nolint comments.
func Report(pass *analysis.Pass, diag analysis.Diagnostic) {
	// Get the file containing this position
	file := findFile(pass, diag.Pos)
	if file == nil {
		pass.Report(diag)
		return
	}

	if ShouldSkip(pass.Fset, file.Comments, diag.Pos, pass.Analyzer.Name) {
		return
	}

	pass.Report(diag)
}

// findFile finds the *ast.File containing the given position.
func findFile(pass *analysis.Pass, pos token.Pos) *ast.File {
	for _, file := range pass.Files {
		if file.Pos() <= pos && pos <= file.End() {
			return file
		}
	}
	return nil
}
