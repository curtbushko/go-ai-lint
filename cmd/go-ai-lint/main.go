// Command go-ai-lint is a static analysis tool for detecting common mistakes
// in AI-generated Go code.
//
// Usage:
//
//	go-ai-lint [flags] [packages]
//
// Run with --help for available flags and subcommands.
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
