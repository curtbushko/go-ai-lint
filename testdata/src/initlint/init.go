package initlint

import (
	"net/http"
	"os"
)

// ===== AIL100: init-with-network =====

// init makes network call - bad practice.
func init() {
	resp, _ := http.Get("http://example.com") // want "AIL100: network call in init"
	_ = resp
}

// ===== AIL101: init-with-file-io =====

// init reads file - bad practice.
func init() {
	data, _ := os.ReadFile("config.json") // want "AIL101: file I/O in init"
	_ = data
}

// init opens file - bad practice.
func init() {
	f, _ := os.Open("file.txt") // want "AIL101: file I/O in init"
	_ = f
}

// init creates file - bad practice.
func init() {
	f, _ := os.Create("output.txt") // want "AIL101: file I/O in init"
	_ = f
}

// GoodInit only does simple initialization.
func init() {
	// Just variable initialization - OK
	globalConfig = defaultConfig()
}

// GoodInitRegister only does registration - common pattern.
func init() {
	// Registration is a common pattern - OK
	registerHandler()
}

// Helper variables and functions.
var globalConfig string

func defaultConfig() string { return "default" }
func registerHandler()      {}
