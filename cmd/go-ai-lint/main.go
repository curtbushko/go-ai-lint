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
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/curtbushko/go-ai-lint/internal/application/concurrencylint"
	"github.com/curtbushko/go-ai-lint/internal/application/contextlint"
	"github.com/curtbushko/go-ai-lint/internal/application/deferlint"
	"github.com/curtbushko/go-ai-lint/internal/application/errorlint"
	"github.com/curtbushko/go-ai-lint/internal/application/goroutinelint"
	"github.com/curtbushko/go-ai-lint/internal/application/initlint"
	"github.com/curtbushko/go-ai-lint/internal/application/interfacelint"
	"github.com/curtbushko/go-ai-lint/internal/application/naminglint"
	"github.com/curtbushko/go-ai-lint/internal/application/optionlint"
	"github.com/curtbushko/go-ai-lint/internal/application/paniclint"
	"github.com/curtbushko/go-ai-lint/internal/application/slicemaplint"
	"github.com/curtbushko/go-ai-lint/internal/application/stringlint"
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
