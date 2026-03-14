// Package main only calls Execute() - no cobra.Command initialization in main.go.
// AIL122 should NOT flag this file because it only calls Execute(), does not init cobra.Command.
package main

import (
	"os"

	// cobra is imported but not used to create &cobra.Command{}
	_ "github.com/spf13/cobra"
)

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
