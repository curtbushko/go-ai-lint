// Package main is a cmd/main.go with flag usage but no viper import.
package main // want "AIL120: cmd/main.go should use cobra for CLI structure" `AIL121: cmd/\*.go with flag usage should also use viper for configuration`

import (
	"flag"
	"fmt"
)

var debug = flag.Bool("debug", false, "enable debug mode")

func main() {
	flag.Parse()
	if *debug {
		fmt.Println("Debug mode enabled")
	}
	fmt.Println("Hello without viper")
}
