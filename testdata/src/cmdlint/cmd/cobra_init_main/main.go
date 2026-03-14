// Package main is a cmd/main.go that initializes cobra.Command directly in main.go.
// This is a bad practice - cobra.Command should be defined in root.go or similar, not main.go.
package main // want "AIL122: cobra.Command should be initialized in root.go, not main.go"

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is defined in main.go (bad practice)
var rootCmd = &cobra.Command{}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
