package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/config"
)

// Test constants to avoid goconst lint warnings.
const (
	severityHigh = "high"
	formatJSON   = "json"
)

func TestConfigFlagLoadsSpecifiedFile(t *testing.T) {
	// Given: Create temp config file with custom settings
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yml")
	configContent := `
version: 1
output:
  format: sarif
severity:
  min-severity: critical
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// When: Parse CLI args with --config flag
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + configPath})

	// Then: Config is loaded from specified path
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseConfig() returned nil config")
	}
	if cfg.Output.Format != "sarif" {
		t.Errorf("Output.Format = %q, want sarif", cfg.Output.Format)
	}
	if cfg.Severity.MinSeverity != "critical" {
		t.Errorf("Severity.MinSeverity = %q, want critical", cfg.Severity.MinSeverity)
	}
}

func TestConfigFlagNonExistentFileError(t *testing.T) {
	// Given: No config file at path
	nonExistentPath := "/nonexistent/path/config.yml"

	// When: Call CLI with --config pointing to non-existent file
	cli := NewCLI()
	_, err := cli.ParseConfig([]string{"--config=" + nonExistentPath})

	// Then: Clear error message about missing config file
	if err == nil {
		t.Fatal("ParseConfig() expected error for non-existent config file")
	}
}

func TestConfigFlagPrecedence(t *testing.T) {
	// Given: Config file in current dir AND explicit --config pointing elsewhere
	tmpDir := t.TempDir()

	// Create "discovered" config in tmpDir (simulates current directory config)
	discoveredPath := filepath.Join(tmpDir, ".go-ai-lint.yml")
	discoveredContent := `
version: 1
output:
  format: json
`
	if err := os.WriteFile(discoveredPath, []byte(discoveredContent), 0644); err != nil {
		t.Fatalf("failed to write discovered config: %v", err)
	}

	// Create explicit config in a subdirectory
	explicitDir := filepath.Join(tmpDir, "explicit")
	if err := os.MkdirAll(explicitDir, 0755); err != nil {
		t.Fatalf("failed to create explicit dir: %v", err)
	}
	explicitPath := filepath.Join(explicitDir, "my-config.yml")
	explicitContent := `
version: 1
output:
  format: ai
`
	if err := os.WriteFile(explicitPath, []byte(explicitContent), 0644); err != nil {
		t.Fatalf("failed to write explicit config: %v", err)
	}

	// When: Call CLI with --config flag pointing to explicit config
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + explicitPath})

	// Then: Explicit config is used, not discovered one
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.Output.Format != "ai" {
		t.Errorf("Output.Format = %q, want ai (explicit config should take precedence)", cfg.Output.Format)
	}
}

func TestNoConfigFlagUsesDefaultDiscovery(t *testing.T) {
	// Given: No --config flag provided
	cli := NewCLI()

	// When: Parse args without --config
	cfg, err := cli.ParseConfig([]string{})

	// Then: Default config is returned (since no .go-ai-lint.yml in test context)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseConfig() returned nil config")
	}
	// Should have defaults
	if cfg.Output.Format != "text" {
		t.Errorf("Output.Format = %q, want text (default)", cfg.Output.Format)
	}
}

func TestRemainingArgsPassedThrough(t *testing.T) {
	// Given: CLI args with --config and additional analyzer flags
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte("version: 1"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Parse args
	cli := NewCLI()
	_, err := cli.ParseConfig([]string{"--config=" + configPath, "-deferlint", "./..."})

	// Then: Config loaded successfully and remaining args available
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	remaining := cli.RemainingArgs()
	if len(remaining) != 2 {
		t.Errorf("RemainingArgs() len = %d, want 2", len(remaining))
	}
}

func TestShowConfigDisplaysYAML(t *testing.T) {
	// Given: Config file exists with custom settings
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yml")
	configContent := `
version: 1
output:
  format: sarif
severity:
  min-severity: critical
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// When: Call CLI with --show-config
	cli := NewCLI()
	var buf bytes.Buffer
	shouldExit, err := cli.ParseAndExecute([]string{"--config=" + configPath, "--show-config"}, &buf)

	// Then: Config is printed as valid YAML to stdout and program should exit
	if err != nil {
		t.Fatalf("ParseAndExecute() error = %v", err)
	}
	if !shouldExit {
		t.Error("ParseAndExecute() shouldExit = false, want true for --show-config")
	}

	output := buf.String()
	if !strings.Contains(output, "format: sarif") {
		t.Errorf("Output missing 'format: sarif'\nGot:\n%s", output)
	}
	if !strings.Contains(output, "min-severity: critical") {
		t.Errorf("Output missing 'min-severity: critical'\nGot:\n%s", output)
	}
}

func TestShowConfigShowsDefaults(t *testing.T) {
	// Given: No config file in any search path
	tmpDir := t.TempDir()
	// Change to temp dir with no config file
	oldWd, wdErr := os.Getwd()
	if wdErr != nil {
		t.Fatalf("failed to get working directory: %v", wdErr)
	}
	if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
		t.Fatalf("failed to change directory: %v", chdirErr)
	}
	defer func() {
		if restoreErr := os.Chdir(oldWd); restoreErr != nil {
			t.Logf("failed to restore working directory: %v", restoreErr)
		}
	}()

	// When: Call CLI with --show-config
	cli := NewCLI()
	var buf bytes.Buffer
	shouldExit, err := cli.ParseAndExecute([]string{"--show-config"}, &buf)

	// Then: Default config is printed as YAML
	if err != nil {
		t.Fatalf("ParseAndExecute() error = %v", err)
	}
	if !shouldExit {
		t.Error("ParseAndExecute() shouldExit = false, want true for --show-config")
	}

	output := buf.String()
	// Default format is "text"
	if !strings.Contains(output, "format: text") {
		t.Errorf("Output missing 'format: text' (default)\nGot:\n%s", output)
	}
	// Should indicate defaults were used
	if !strings.Contains(output, "# Source: defaults") {
		t.Errorf("Output missing source annotation for defaults\nGot:\n%s", output)
	}
}

func TestShowConfigIncludesSource(t *testing.T) {
	// Given: Config file exists
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "my-config.yml")
	configContent := `version: 1`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// When: Call CLI with --show-config
	cli := NewCLI()
	var buf bytes.Buffer
	shouldExit, err := cli.ParseAndExecute([]string{"--config=" + configPath, "--show-config"}, &buf)

	// Then: Output includes comment with config source path
	if err != nil {
		t.Fatalf("ParseAndExecute() error = %v", err)
	}
	if !shouldExit {
		t.Error("ParseAndExecute() shouldExit = false, want true for --show-config")
	}

	output := buf.String()
	expectedSource := "# Source: " + configPath
	if !strings.Contains(output, expectedSource) {
		t.Errorf("Output missing source annotation %q\nGot:\n%s", expectedSource, output)
	}
}

func TestInitCreatesConfigFile(t *testing.T) {
	// Given: No .go-ai-lint.yml exists in working directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, config.ConfigFileName)

	// When: Call CLI with --init
	cli := NewCLI()
	var buf bytes.Buffer
	shouldExit, err := cli.ParseAndExecute([]string{"--init", "--dir=" + tmpDir}, &buf)

	// Then: .go-ai-lint.yml is created with default values and comments
	if err != nil {
		t.Fatalf("ParseAndExecute() error = %v", err)
	}
	if !shouldExit {
		t.Error("ParseAndExecute() shouldExit = false, want true for --init")
	}

	// Verify file was created
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Verify content contains expected sections
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config file: %v", readErr)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "version: 1") {
		t.Errorf("Config file missing 'version: 1'\nGot:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "# go-ai-lint configuration") {
		t.Errorf("Config file missing header comment\nGot:\n%s", contentStr)
	}

	// Verify success message was printed
	output := buf.String()
	if !strings.Contains(output, configPath) {
		t.Errorf("Output missing config path\nGot:\n%s", output)
	}
}

func TestInitRefusesOverwrite(t *testing.T) {
	// Given: .go-ai-lint.yml already exists
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, config.ConfigFileName)
	existingContent := "version: 1\noutput:\n  format: json\n"
	if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}

	// When: Call CLI with --init
	cli := NewCLI()
	var buf bytes.Buffer
	_, err := cli.ParseAndExecute([]string{"--init", "--dir=" + tmpDir}, &buf)

	// Then: Error message about existing config, file not modified
	if err == nil {
		t.Fatal("ParseAndExecute() expected error for existing config")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Error should mention 'already exists', got: %v", err)
	}

	// Verify file was not modified
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config file: %v", readErr)
	}
	if string(content) != existingContent {
		t.Errorf("Config file was modified when it should not have been")
	}
}

func TestInitForceOverwrites(t *testing.T) {
	// Given: .go-ai-lint.yml already exists with custom settings
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, config.ConfigFileName)
	existingContent := "version: 1\noutput:\n  format: json\n"
	if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}

	// When: Call CLI with --init --force
	cli := NewCLI()
	var buf bytes.Buffer
	shouldExit, err := cli.ParseAndExecute([]string{"--init", "--force", "--dir=" + tmpDir}, &buf)

	// Then: Config file is replaced with defaults
	if err != nil {
		t.Fatalf("ParseAndExecute() error = %v", err)
	}
	if !shouldExit {
		t.Error("ParseAndExecute() shouldExit = false, want true for --init")
	}

	// Verify file was overwritten with new content
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config file: %v", readErr)
	}
	contentStr := string(content)

	// Should have default template content, not the old json format
	if strings.Contains(contentStr, "format: json") {
		t.Errorf("Config file should have been overwritten but still has old content")
	}
	if !strings.Contains(contentStr, "# go-ai-lint configuration") {
		t.Errorf("Config file missing header comment after overwrite\nGot:\n%s", contentStr)
	}
}

func TestEnableFlagAddsAnalyzer(t *testing.T) {
	// Given: Config has enable-all: false
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: 1
analyzers:
  enable-all: false
  enable: []
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with --enable=deferlint,errorlint
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--enable=deferlint,errorlint"})

	// Then: deferlint and errorlint are enabled, others disabled
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if !cfg.IsAnalyzerEnabled("deferlint") {
		t.Error("deferlint should be enabled after --enable flag")
	}
	if !cfg.IsAnalyzerEnabled("errorlint") {
		t.Error("errorlint should be enabled after --enable flag")
	}
	if cfg.IsAnalyzerEnabled("optionlint") {
		t.Error("optionlint should remain disabled (not in --enable list)")
	}
}

func TestDisableFlagRemovesAnalyzer(t *testing.T) {
	// Given: Config has enable-all: true
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: 1
analyzers:
  enable-all: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with --disable=optionlint
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--disable=optionlint"})

	// Then: optionlint is disabled, all others remain enabled
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.IsAnalyzerEnabled("optionlint") {
		t.Error("optionlint should be disabled after --disable flag")
	}
	if !cfg.IsAnalyzerEnabled("deferlint") {
		t.Error("deferlint should remain enabled (enable-all: true)")
	}
}

func TestMinSeverityFlag(t *testing.T) {
	// Given: Config has min-severity: low
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: 1
severity:
  min-severity: low
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with --min-severity=high
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--min-severity=" + severityHigh})

	// Then: min-severity is high
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.Severity.MinSeverity != severityHigh {
		t.Errorf("Severity.MinSeverity = %q, want %s", cfg.Severity.MinSeverity, severityHigh)
	}
}

func TestFormatFlag(t *testing.T) {
	// Given: Config has format: text
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: 1
output:
  format: text
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with --format=json
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--format=" + formatJSON})

	// Then: Output format is JSON
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.Output.Format != formatJSON {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
	}
}

func TestCLIFlagsOverrideConfig(t *testing.T) {
	// Given: Config file has format: text, min-severity: low
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: 1
output:
  format: text
severity:
  min-severity: low
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with --format=json --min-severity=high
	cli := NewCLI()
	cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--format=" + formatJSON, "--min-severity=" + severityHigh})

	// Then: Merged config uses CLI values
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.Output.Format != formatJSON {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
	}
	if cfg.Severity.MinSeverity != severityHigh {
		t.Errorf("Severity.MinSeverity = %q, want %s", cfg.Severity.MinSeverity, severityHigh)
	}
}

func TestInvalidMinSeverityError(t *testing.T) {
	// Given: Config file exists
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte("version: 1"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with invalid --min-severity
	cli := NewCLI()
	_, err := cli.ParseConfig([]string{"--config=" + configPath, "--min-severity=invalid"})

	// Then: Clear error about invalid severity
	if err == nil {
		t.Fatal("ParseConfig() expected error for invalid min-severity")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Error should mention 'invalid', got: %v", err)
	}
}

func TestInvalidFormatError(t *testing.T) {
	// Given: Config file exists
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte("version: 1"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: Call with invalid --format
	cli := NewCLI()
	_, err := cli.ParseConfig([]string{"--config=" + configPath, "--format=invalid"})

	// Then: Clear error about invalid format
	if err == nil {
		t.Fatal("ParseConfig() expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Error should mention 'invalid', got: %v", err)
	}
}

func TestAllValidFormats(t *testing.T) {
	validFormats := []string{"text", formatJSON, "ai", "sarif"}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yml")
			if err := os.WriteFile(configPath, []byte("version: 1"), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			cli := NewCLI()
			cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--format=" + format})

			if err != nil {
				t.Fatalf("ParseConfig() error = %v for format %s", err, format)
			}
			if cfg.Output.Format != format {
				t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, format)
			}
		})
	}
}

func TestAllValidSeverities(t *testing.T) {
	validSeverities := []string{"low", "medium", severityHigh, "critical"}

	for _, severity := range validSeverities {
		t.Run(severity, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yml")
			if err := os.WriteFile(configPath, []byte("version: 1"), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			cli := NewCLI()
			cfg, err := cli.ParseConfig([]string{"--config=" + configPath, "--min-severity=" + severity})

			if err != nil {
				t.Fatalf("ParseConfig() error = %v for severity %s", err, severity)
			}
			if cfg.Severity.MinSeverity != severity {
				t.Errorf("Severity.MinSeverity = %q, want %s", cfg.Severity.MinSeverity, severity)
			}
		})
	}
}

func TestConfigMethodReturnsLoadedConfig(t *testing.T) {
	// Given: Config file with custom settings
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: 1
nolint:
  enabled: false
  require-specific: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When: ParseAndExecute is called
	cli := NewCLI()
	var buf bytes.Buffer
	shouldExit, err := cli.ParseAndExecute([]string{"--config=" + configPath}, &buf)

	// Then: Config() returns the loaded config
	if err != nil {
		t.Fatalf("ParseAndExecute() error = %v", err)
	}
	if shouldExit {
		t.Error("ParseAndExecute() shouldExit = true, want false for normal operation")
	}

	cfg := cli.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil after ParseAndExecute")
	}
	if cfg.Nolint.Enabled {
		t.Error("Config().Nolint.Enabled = true, want false")
	}
	if !cfg.Nolint.RequireSpecific {
		t.Error("Config().Nolint.RequireSpecific = false, want true")
	}
}

func TestConfigMethodReturnsNilBeforeParse(t *testing.T) {
	// Given: New CLI without parsing
	cli := NewCLI()

	// When: Config() is called before ParseAndExecute
	cfg := cli.Config()

	// Then: Returns nil
	if cfg != nil {
		t.Error("Config() should return nil before ParseAndExecute")
	}
}
