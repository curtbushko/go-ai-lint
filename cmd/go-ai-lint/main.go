// Command go-ai-lint is a static analysis tool for detecting common mistakes
// in AI-generated Go code.
//
// Usage:
//
//	go-ai-lint [flags] [packages]
//
// Run with -help for available flags.
package main

import (
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/concurrencylint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/contextlint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/deferlint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/errorlint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/goroutinelint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/initlint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/interfacelint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/naminglint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/optionlint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/paniclint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/slicemaplint"
	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/stringlint"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		deferlint.New().Analyzer(),
		contextlint.New().Analyzer(),
		goroutinelint.New().Analyzer(),
		slicemaplint.New().Analyzer(),
		errorlint.New().Analyzer(),
		concurrencylint.New().Analyzer(),
		naminglint.New().Analyzer(),
		interfacelint.New().Analyzer(),
		stringlint.New().Analyzer(),
		paniclint.New().Analyzer(),
		initlint.New().Analyzer(),
		optionlint.New().Analyzer(),
	)
}
