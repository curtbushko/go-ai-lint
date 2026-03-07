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
	"fmt"
	"os"

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
	// Parse --config flag before delegating to multichecker
	cli := NewCLI()
	_, err := cli.ParseConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-ai-lint: %v\n", err)
		os.Exit(1)
	}

	// Replace os.Args with remaining args for multichecker
	// multichecker.Main reads os.Args directly
	os.Args = append([]string{os.Args[0]}, cli.RemainingArgs()...)

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
