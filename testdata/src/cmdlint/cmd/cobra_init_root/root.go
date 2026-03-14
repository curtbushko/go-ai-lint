// Package main defines the root command in root.go (good practice).
// AIL122 should NOT flag this file because it's not main.go.
package main

import (
	"github.com/spf13/cobra"
)

// rootCmd is properly defined in root.go (good practice)
var rootCmd = &cobra.Command{}

func Execute() error {
	return rootCmd.Execute()
}
