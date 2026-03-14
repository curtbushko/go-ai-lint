// Package main is a cmd/main.go with viper import (good case for AIL121).
// Note: This will still trigger AIL120 (missing cobra) but not AIL121 (has viper).
package main // want "AIL120: cmd/main.go should use cobra for CLI structure"

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

var debug = flag.Bool("debug", false, "enable debug mode")

func main() {
	flag.Parse()
	viper.AutomaticEnv()
	if *debug {
		fmt.Println("Debug mode enabled")
	}
	fmt.Println("Hello with viper")
}
