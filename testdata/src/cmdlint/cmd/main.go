// Package main is a cmd/main.go without cobra import.
package main // want "AIL120: cmd/main.go should use cobra for CLI structure"

import (
	"fmt"
)

func main() {
	fmt.Println("Hello without cobra")
}
