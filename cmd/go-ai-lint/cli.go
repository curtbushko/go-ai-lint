// Package main provides the CLI wrapper for go-ai-lint.
// It handles --config flag parsing before delegating to the multichecker.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/curtbushko/go-ai-lint/internal/config"
)

// CLI handles command-line argument parsing for go-ai-lint.
type CLI struct {
	configPath    string
	showConfig    bool
	initConfig    bool
	forceInit     bool
	initDir       string // Directory for --init (defaults to cwd)
	enable        string // --enable flag: comma-separated analyzer names
	disable       string // --disable flag: comma-separated analyzer names
	minSeverity   string // --min-severity flag: low, medium, high, critical
	format        string // --format flag: text, json, ai, sarif
	remainingArgs []string
	flagSet       *flag.FlagSet
	configSource  string         // Tracks where config was loaded from
	loadedConfig  *config.Config // Loaded config after ParseAndExecute
}

// NewCLI creates a new CLI instance.
func NewCLI() *CLI {
	return &CLI{
		flagSet: flag.NewFlagSet("go-ai-lint", flag.ContinueOnError),
	}
}

// ParseConfig parses command-line arguments and loads configuration.
// It extracts the --config flag and loads config from the specified path,
// or uses default discovery if no --config is provided.
func (c *CLI) ParseConfig(args []string) (*config.Config, error) {
	c.flagSet.StringVar(&c.configPath, "config", "", "path to configuration file")
	c.flagSet.BoolVar(&c.showConfig, "show-config", false, "display resolved configuration and exit")
	c.flagSet.BoolVar(&c.initConfig, "init", false, "generate default config file in current directory")
	c.flagSet.BoolVar(&c.forceInit, "force", false, "overwrite existing config when using --init")
	c.flagSet.StringVar(&c.initDir, "dir", "", "directory for --init (defaults to current directory)")
	c.flagSet.StringVar(&c.enable, "enable", "", "comma-separated list of analyzers to enable")
	c.flagSet.StringVar(&c.disable, "disable", "", "comma-separated list of analyzers to disable")
	c.flagSet.StringVar(&c.minSeverity, "min-severity", "", "minimum severity to report: low, medium, high, critical")
	c.flagSet.StringVar(&c.format, "format", "", "output format: text, json, ai, sarif")

	// Parse only our flags, leave other flags for multichecker
	ourArgs, remaining := c.separateArgs(args)
	c.remainingArgs = remaining

	// Parse our args
	if err := c.flagSet.Parse(ourArgs); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	// For --init, we don't need to load existing config
	if c.initConfig {
		return nil, nil
	}

	cfg, err := c.loadConfig()
	if err != nil {
		return nil, err
	}

	// Build CLI overrides and merge with loaded config
	overrides, err := c.buildOverrides()
	if err != nil {
		return nil, err
	}
	cfg.Merge(overrides)

	return cfg, nil
}

// flagSpec defines how to recognize and parse a CLI flag.
type flagSpec struct {
	name     string // flag name without dashes (e.g., "config")
	hasValue bool   // whether the flag takes a value
	boolFlag bool   // whether this is a boolean flag (no value)
}

// knownFlags lists all flags that should be captured by the CLI.
var knownFlags = []flagSpec{
	{name: "config", hasValue: true},
	{name: "show-config", boolFlag: true},
	{name: "init", boolFlag: true},
	{name: "force", boolFlag: true},
	{name: "dir", hasValue: true},
	{name: "enable", hasValue: true},
	{name: "disable", hasValue: true},
	{name: "min-severity", hasValue: true},
	{name: "format", hasValue: true},
}

// separateArgs separates CLI args into our flags and remaining args for multichecker.
func (c *CLI) separateArgs(args []string) (ourArgs, remaining []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		matched, skip := matchFlag(arg, args, i)
		if matched != "" {
			ourArgs = append(ourArgs, matched)
			if skip > 0 {
				i++
				ourArgs = append(ourArgs, args[i])
			}
		} else {
			remaining = append(remaining, arg)
		}
	}
	return ourArgs, remaining
}

// matchFlag checks if arg matches any known flag and returns the matched arg and skip count.
// Returns empty string if no match. Returns skip=1 if next arg should be consumed as value.
func matchFlag(arg string, args []string, idx int) (matched string, skip int) {
	for _, spec := range knownFlags {
		if matched, skip := matchValueFlag(arg, args, idx, spec); matched != "" {
			return matched, skip
		}
		if matched := matchBoolFlag(arg, spec); matched != "" {
			return matched, 0
		}
	}
	return "", 0
}

// matchValueFlag checks if arg matches a value-taking flag spec.
func matchValueFlag(arg string, args []string, idx int, spec flagSpec) (string, int) {
	if !spec.hasValue {
		return "", 0
	}
	// Check --flag=value or -flag=value
	if strings.HasPrefix(arg, "--"+spec.name+"=") || strings.HasPrefix(arg, "-"+spec.name+"=") {
		return arg, 0
	}
	// Check --flag or -flag (value follows)
	if arg == "--"+spec.name || arg == "-"+spec.name {
		if idx+1 < len(args) {
			return arg, 1
		}
		return arg, 0
	}
	return "", 0
}

// matchBoolFlag checks if arg matches a boolean flag spec.
func matchBoolFlag(arg string, spec flagSpec) string {
	if !spec.boolFlag {
		return ""
	}
	if arg == "--"+spec.name || arg == "-"+spec.name {
		return arg
	}
	return ""
}

// loadConfig loads the configuration from the appropriate source.
func (c *CLI) loadConfig() (*config.Config, error) {
	if c.configPath != "" {
		return c.loadFromExplicitPath()
	}
	return c.loadFromDiscovery()
}

// loadFromExplicitPath loads config from the explicitly specified path.
func (c *CLI) loadFromExplicitPath() (*config.Config, error) {
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", c.configPath)
	}
	c.configSource = c.configPath
	return config.LoadFromPath(c.configPath)
}

// loadFromDiscovery loads config using default discovery from current directory.
func (c *CLI) loadFromDiscovery() (*config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return nil, err
	}

	c.configSource = config.FindConfigFile(cwd)
	if c.configSource == "" {
		c.configSource = "defaults"
	}

	return cfg, nil
}

// ParseAndExecute parses arguments and executes any immediate actions like --show-config.
// Returns (shouldExit, error) where shouldExit indicates if the program should exit
// after this call (e.g., after displaying config).
func (c *CLI) ParseAndExecute(args []string, w io.Writer) (bool, error) {
	cfg, err := c.ParseConfig(args)
	if err != nil {
		return false, err
	}

	// Store the loaded config for later access
	c.loadedConfig = cfg

	if c.initConfig {
		if err := c.executeInit(w); err != nil {
			return false, err
		}
		return true, nil
	}

	if c.showConfig {
		if err := c.writeConfig(cfg, w); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}

// Config returns the loaded configuration after ParseAndExecute.
// Returns nil if ParseAndExecute hasn't been called or returned an error.
func (c *CLI) Config() *config.Config {
	return c.loadedConfig
}

// executeInit handles the --init flag logic.
func (c *CLI) executeInit(w io.Writer) error {
	// Determine target directory
	targetDir := c.initDir
	if targetDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		targetDir = cwd
	}

	configPath := filepath.Join(targetDir, config.ConfigFileName)

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		if !c.forceInit {
			return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
		}
	}

	// Generate and write default config
	content := config.GenerateDefaultConfig()
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	// Print success message
	if _, err := fmt.Fprintf(w, "Created config file: %s\n", configPath); err != nil {
		return fmt.Errorf("write success message: %w", err)
	}

	return nil
}

// writeConfig writes the configuration to the given writer with source annotation.
func (c *CLI) writeConfig(cfg *config.Config, w io.Writer) error {
	yamlStr, err := cfg.ToYAML()
	if err != nil {
		return fmt.Errorf("serialize config: %w", err)
	}

	// Write source annotation
	if _, err := fmt.Fprintf(w, "# Source: %s\n", c.configSource); err != nil {
		return fmt.Errorf("write source annotation: %w", err)
	}
	if _, err := fmt.Fprint(w, yamlStr); err != nil {
		return fmt.Errorf("write config yaml: %w", err)
	}

	return nil
}

// RemainingArgs returns the arguments that were not consumed by the CLI.
// These should be passed to the multichecker.
func (c *CLI) RemainingArgs() []string {
	return c.remainingArgs
}

// ConfigPath returns the explicit config path if one was provided.
func (c *CLI) ConfigPath() string {
	return c.configPath
}

// buildOverrides constructs CLIOverrides from parsed CLI flags.
// It validates flag values and returns an error for invalid inputs.
func (c *CLI) buildOverrides() (config.CLIOverrides, error) {
	overrides := config.CLIOverrides{}

	// Parse --enable flag
	if c.enable != "" {
		overrides.Enable = parseCommaSeparated(c.enable)
	}

	// Parse --disable flag
	if c.disable != "" {
		overrides.Disable = parseCommaSeparated(c.disable)
	}

	// Validate and set --min-severity
	if c.minSeverity != "" {
		if !isValidSeverity(c.minSeverity) {
			return config.CLIOverrides{}, fmt.Errorf("invalid min-severity %q: must be one of low, medium, high, critical", c.minSeverity)
		}
		overrides.MinSeverity = c.minSeverity
	}

	// Validate and set --format
	if c.format != "" {
		if !isValidFormat(c.format) {
			return config.CLIOverrides{}, fmt.Errorf("invalid format %q: must be one of text, json, ai, sarif", c.format)
		}
		overrides.Format = c.format
	}

	return overrides, nil
}

// parseCommaSeparated splits a comma-separated string into a slice.
// Empty strings between commas are ignored.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// isValidSeverity checks if the severity value is valid.
func isValidSeverity(s string) bool {
	switch s {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

// isValidFormat checks if the format value is valid.
func isValidFormat(s string) bool {
	switch s {
	case "text", "json", "ai", "sarif":
		return true
	default:
		return false
	}
}
