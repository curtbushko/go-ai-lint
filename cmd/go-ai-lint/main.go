// Command go-ai-lint is a static analysis tool for detecting common mistakes
// in AI-generated Go code.
//
// Usage:
//
//	go-ai-lint [flags] [packages]
//
// Run with --help for available flags and subcommands.
//
//nolint:cmdlint // cobra is in root.go following standard project layout
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "go-ai-lint: %v\n", err)
		os.Exit(1)
	}
}
