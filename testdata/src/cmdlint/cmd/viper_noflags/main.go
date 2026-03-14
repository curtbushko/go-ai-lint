// Package main is a cmd/main.go without flag usage (edge case - should not trigger AIL121).
// Note: This will still trigger AIL120 (missing cobra) but not AIL121 (no flags).
package main // want "AIL120: cmd/main.go should use cobra for CLI structure"

import (
	"fmt"
)

func main() {
	fmt.Println("Hello without flags or viper")
}
