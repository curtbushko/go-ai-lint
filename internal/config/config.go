// Package config provides configuration loading and parsing for go-ai-lint.
// It supports loading configuration from .go-ai-lint.yml files with precedence:
// 1. Explicit config path (--config flag)
// 2. .go-ai-lint.yml in current directory
// 3. .go-ai-lint.yml in parent directories (up to root)
// 4. Built-in defaults
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigFileName is the default configuration file name.
const ConfigFileName = ".go-ai-lint.yml"

// Valid output formats.
var validOutputFormats = map[string]bool{
	"text":  true,
	"json":  true,
	"ai":    true,
	"sarif": true,
}

// Valid severity levels.
var validSeverities = map[string]bool{
	"critical": true,
	"high":     true,
	"medium":   true,
	"low":      true,
}

// Config represents the go-ai-lint configuration.
type Config struct {
	Version   int            `yaml:"version"`
	Run       RunConfig      `yaml:"run"`
	Output    OutputConfig   `yaml:"output"`
	Nolint    NolintConfig   `yaml:"nolint"`
	Analyzers AnalyzerConfig `yaml:"analyzers"`
	Severity  SeverityConfig `yaml:"severity"`
}

// RunConfig contains runtime settings.
type RunConfig struct {
	Timeout     time.Duration `yaml:"timeout"`
	Concurrency int           `yaml:"concurrency"`
	SkipDirs    []string      `yaml:"skip-dirs"`
	SkipFiles   []string      `yaml:"skip-files"`
}

// OutputConfig contains output settings.
type OutputConfig struct {
	Format            string `yaml:"format"`
	PrintAnalyzerName bool   `yaml:"print-analyzer-name"`
	SortBy            string `yaml:"sort-by"`
}

// NolintConfig contains nolint directive settings.
type NolintConfig struct {
	Enabled         bool `yaml:"enabled"`
	RequireSpecific bool `yaml:"require-specific"`
}

// AnalyzerConfig contains per-analyzer settings.
type AnalyzerConfig struct {
	EnableAll bool     `yaml:"enable-all"`
	Enable    []string `yaml:"enable"`
	Disable   []string `yaml:"disable"`
}

// SeverityConfig contains severity settings.
type SeverityConfig struct {
	MinSeverity string   `yaml:"min-severity"`
	ErrorOn     []string `yaml:"error-on"`
}

// CLIOverrides represents command-line flag overrides for configuration.
// Empty strings and nil slices indicate no override (preserve original value).
type CLIOverrides struct {
	Enable      []string // --enable flag: comma-separated analyzer names to enable
	Disable     []string // --disable flag: comma-separated analyzer names to disable
	MinSeverity string   // --min-severity flag: low, medium, high, critical
	Format      string   // --format flag: text, json, ai, sarif
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Version: 1,
		Run: RunConfig{
			Timeout:     5 * time.Minute,
			Concurrency: 0, // 0 means auto (use runtime.NumCPU)
			SkipDirs:    []string{},
			SkipFiles:   []string{},
		},
		Output: OutputConfig{
			Format:            "text",
			PrintAnalyzerName: true,
			SortBy:            "file",
		},
		Nolint: NolintConfig{
			Enabled:         true,
			RequireSpecific: false,
		},
		Analyzers: AnalyzerConfig{
			EnableAll: true,
			Enable:    []string{},
			Disable:   []string{},
		},
		Severity: SeverityConfig{
			MinSeverity: "low",
			ErrorOn:     []string{},
		},
	}
}

// Load loads configuration starting from the given directory.
// It searches for .go-ai-lint.yml in the directory and parent directories.
// Returns default config if no config file is found.
func Load(startDir string) (*Config, error) {
	configPath := FindConfigFile(startDir)
	if configPath == "" {
		return Default(), nil
	}

	return LoadFromPath(configPath)
}

// LoadFromPath loads configuration from a specific file path.
func LoadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	return LoadFromReader(data)
}

// LoadFromReader loads configuration from YAML bytes.
// Unspecified fields are set to defaults.
func LoadFromReader(data []byte) (*Config, error) {
	cfg := Default()

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// FindConfigFile searches for a config file starting from startDir
// and walking up to parent directories. Returns empty string if not found.
func FindConfigFile(startDir string) string {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}

	for {
		configPath := filepath.Join(absDir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			// Reached root
			break
		}
		absDir = parent
	}

	return ""
}

// ToYAML serializes the configuration to YAML format.
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("marshal config to YAML: %w", err)
	}
	return string(data), nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Output.Format != "" && !validOutputFormats[c.Output.Format] {
		return fmt.Errorf("invalid output format %q: must be one of text, json, ai, sarif", c.Output.Format)
	}

	if c.Severity.MinSeverity != "" && !validSeverities[c.Severity.MinSeverity] {
		return fmt.Errorf("invalid min-severity %q: must be one of critical, high, medium, low", c.Severity.MinSeverity)
	}

	for _, sev := range c.Severity.ErrorOn {
		if !validSeverities[sev] {
			return fmt.Errorf("invalid severity in error-on %q: must be one of critical, high, medium, low", sev)
		}
	}

	return nil
}

// IsAnalyzerEnabled returns whether the given analyzer is enabled.
func (c *Config) IsAnalyzerEnabled(name string) bool {
	// Check if explicitly disabled
	for _, disabled := range c.Analyzers.Disable {
		if disabled == name {
			return false
		}
	}

	// If EnableAll is true, analyzer is enabled (unless in disable list above)
	if c.Analyzers.EnableAll {
		return true
	}

	// If EnableAll is false, check if in enable list
	for _, enabled := range c.Analyzers.Enable {
		if enabled == name {
			return true
		}
	}

	return false
}

// Merge applies CLI overrides to the configuration.
// Empty strings and nil slices in overrides are ignored (preserve original values).
func (c *Config) Merge(overrides CLIOverrides) {
	// Apply enable overrides - add to enable list
	if len(overrides.Enable) > 0 {
		c.Analyzers.Enable = append(c.Analyzers.Enable, overrides.Enable...)
	}

	// Apply disable overrides - add to disable list
	if len(overrides.Disable) > 0 {
		c.Analyzers.Disable = append(c.Analyzers.Disable, overrides.Disable...)
	}

	// Apply min-severity override
	if overrides.MinSeverity != "" {
		c.Severity.MinSeverity = overrides.MinSeverity
	}

	// Apply format override
	if overrides.Format != "" {
		c.Output.Format = overrides.Format
	}
}

// LoadWithOverrides loads configuration with an optional explicit config path.
// If explicitPath is provided, it loads from that path and returns an error if not found.
// If explicitPath is empty, it falls back to standard discovery from startDir.
func LoadWithOverrides(startDir, explicitPath string) (*Config, error) {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", explicitPath)
		}
		return LoadFromPath(explicitPath)
	}
	return Load(startDir)
}

// GenerateDefaultConfig returns a documented YAML configuration template
// with default values and helpful comments.
func GenerateDefaultConfig() string {
	return `# go-ai-lint configuration
# https://github.com/curtbushko/go-ai-lint

version: 1

# Runtime settings
run:
  # Maximum time to wait for analysis to complete
  timeout: 5m
  # Number of concurrent analyzers (0 = auto, uses runtime.NumCPU)
  concurrency: 0
  # Directories to skip during analysis
  skip-dirs: []
  # File patterns to skip during analysis
  skip-files: []

# Output settings
output:
  # Output format: text, json, ai, sarif
  format: text
  # Include analyzer name in output
  print-analyzer-name: true
  # Sort results by: file, severity
  sort-by: file

# Nolint directive settings
nolint:
  # Enable processing of //nolint directives
  enabled: true
  # Require specific analyzer name in //nolint directive
  require-specific: false

# Analyzer settings
analyzers:
  # Enable all analyzers by default
  enable-all: true
  # Explicitly enable specific analyzers (when enable-all is false)
  enable: []
  # Disable specific analyzers
  disable: []

# Severity settings
severity:
  # Minimum severity to report: critical, high, medium, low
  min-severity: low
  # Severities that cause non-zero exit code
  error-on: []
`
}
