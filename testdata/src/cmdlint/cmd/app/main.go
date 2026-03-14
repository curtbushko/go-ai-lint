// Package main is a cmd/app/main.go with cobra import.
// This passes AIL120 (has cobra) but fails AIL122 (initializes cobra.Command in main.go).
package main // want "AIL122: cobra.Command should be initialized in root.go, not main.go"

import (
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{}
	_ = cmd.Execute()
}
