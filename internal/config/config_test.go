package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/curtbushko/go-ai-lint/internal/config"
)

// Test format constants to avoid goconst lint warnings.
const (
	formatJSON  = "json"
	formatText  = "text"
	formatAI    = "ai"
	formatSarif = "sarif"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()

	// Check version
	if cfg.Version != 1 {
		t.Errorf("Default().Version = %d, want 1", cfg.Version)
	}

	// Check run defaults
	if cfg.Run.Timeout != 5*time.Minute {
		t.Errorf("Default().Run.Timeout = %v, want 5m", cfg.Run.Timeout)
	}
	if cfg.Run.Concurrency != 0 {
		t.Errorf("Default().Run.Concurrency = %d, want 0 (auto)", cfg.Run.Concurrency)
	}

	// Check output defaults
	if cfg.Output.Format != "text" {
		t.Errorf("Default().Output.Format = %q, want %q", cfg.Output.Format, "text")
	}
	if !cfg.Output.PrintAnalyzerName {
		t.Error("Default().Output.PrintAnalyzerName = false, want true")
	}
	if cfg.Output.SortBy != "file" {
		t.Errorf("Default().Output.SortBy = %q, want %q", cfg.Output.SortBy, "file")
	}

	// Check nolint defaults
	if !cfg.Nolint.Enabled {
		t.Error("Default().Nolint.Enabled = false, want true")
	}
	if cfg.Nolint.RequireSpecific {
		t.Error("Default().Nolint.RequireSpecific = true, want false")
	}

	// Check analyzer defaults
	if !cfg.Analyzers.EnableAll {
		t.Error("Default().Analyzers.EnableAll = false, want true")
	}

	// Check severity defaults
	if cfg.Severity.MinSeverity != "low" {
		t.Errorf("Default().Severity.MinSeverity = %q, want %q", cfg.Severity.MinSeverity, "low")
	}
}

func TestLoadFromReader(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		check   func(t *testing.T, cfg *config.Config)
		wantErr bool
	}{
		{
			name: "valid minimal config",
			yaml: `version: 1`,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Version != 1 {
					t.Errorf("Version = %d, want 1", cfg.Version)
				}
			},
			wantErr: false,
		},
		{
			name: "config with run settings",
			yaml: `
version: 1
run:
  timeout: 10m
  concurrency: 4
  skip-dirs:
    - vendor
    - testdata
  skip-files:
    - ".*_mock.go"
`,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Run.Timeout != 10*time.Minute {
					t.Errorf("Run.Timeout = %v, want 10m", cfg.Run.Timeout)
				}
				if cfg.Run.Concurrency != 4 {
					t.Errorf("Run.Concurrency = %d, want 4", cfg.Run.Concurrency)
				}
				if len(cfg.Run.SkipDirs) != 2 {
					t.Errorf("Run.SkipDirs length = %d, want 2", len(cfg.Run.SkipDirs))
				}
				if len(cfg.Run.SkipFiles) != 1 {
					t.Errorf("Run.SkipFiles length = %d, want 1", len(cfg.Run.SkipFiles))
				}
			},
			wantErr: false,
		},
		{
			name: "config with output settings",
			yaml: `
version: 1
output:
  format: json
  print-analyzer-name: false
  sort-by: severity
`,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Output.Format != formatJSON {
					t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
				}
				if cfg.Output.PrintAnalyzerName {
					t.Error("Output.PrintAnalyzerName = true, want false")
				}
				if cfg.Output.SortBy != "severity" {
					t.Errorf("Output.SortBy = %q, want severity", cfg.Output.SortBy)
				}
			},
			wantErr: false,
		},
		{
			name: "config with nolint settings",
			yaml: `
version: 1
nolint:
  enabled: false
  require-specific: true
`,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Nolint.Enabled {
					t.Error("Nolint.Enabled = true, want false")
				}
				if !cfg.Nolint.RequireSpecific {
					t.Error("Nolint.RequireSpecific = false, want true")
				}
			},
			wantErr: false,
		},
		{
			name: "config with analyzer settings",
			yaml: `
version: 1
analyzers:
  enable-all: false
  disable:
    - optionlint
    - stringlint
`,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Analyzers.EnableAll {
					t.Error("Analyzers.EnableAll = true, want false")
				}
				if len(cfg.Analyzers.Disable) != 2 {
					t.Errorf("Analyzers.Disable length = %d, want 2", len(cfg.Analyzers.Disable))
				}
			},
			wantErr: false,
		},
		{
			name: "config with severity settings",
			yaml: `
version: 1
severity:
  min-severity: medium
  error-on:
    - critical
    - high
`,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Severity.MinSeverity != "medium" {
					t.Errorf("Severity.MinSeverity = %q, want medium", cfg.Severity.MinSeverity)
				}
				if len(cfg.Severity.ErrorOn) != 2 {
					t.Errorf("Severity.ErrorOn length = %d, want 2", len(cfg.Severity.ErrorOn))
				}
			},
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			yaml:    `version: [invalid`,
			check:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromReader([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && err == nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestLoadConfigFromCurrentDir(t *testing.T) {
	// Create a temp directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".go-ai-lint.yml")

	configContent := `
version: 1
output:
  format: json
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config starting from tmpDir
	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Output.Format != formatJSON {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
	}
}

func TestLoadConfigFromParentDir(t *testing.T) {
	// Create a temp directory structure: parent/.go-ai-lint.yml and parent/child/
	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "parent")
	childDir := filepath.Join(parentDir, "child")

	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	configPath := filepath.Join(parentDir, ".go-ai-lint.yml")
	configContent := `
version: 1
output:
  format: ai
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config starting from childDir (should find in parentDir)
	cfg, err := config.Load(childDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Output.Format != "ai" {
		t.Errorf("Output.Format = %q, want ai", cfg.Output.Format)
	}
}

func TestLoadConfigReturnsDefaultsWhenNoConfigFound(t *testing.T) {
	// Create a temp directory with no config file
	tmpDir := t.TempDir()

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should return defaults
	defaultCfg := config.Default()
	if cfg.Version != defaultCfg.Version {
		t.Errorf("Version = %d, want %d", cfg.Version, defaultCfg.Version)
	}
	if cfg.Output.Format != defaultCfg.Output.Format {
		t.Errorf("Output.Format = %q, want %q", cfg.Output.Format, defaultCfg.Output.Format)
	}
}

func TestLoadConfigFromExplicitPath(t *testing.T) {
	// Create a temp config file in an unusual location
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yml")

	configContent := `
version: 1
severity:
  min-severity: critical
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if cfg.Severity.MinSeverity != "critical" {
		t.Errorf("Severity.MinSeverity = %q, want critical", cfg.Severity.MinSeverity)
	}
}

func TestLoadConfigFromExplicitPathNotFound(t *testing.T) {
	_, err := config.LoadFromPath("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("LoadFromPath() expected error for non-existent file")
	}
}

func TestConfigMergeWithDefaults(t *testing.T) {
	// Load a partial config - should have defaults for unspecified fields
	yaml := `
version: 1
output:
  format: sarif
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// Explicitly set field
	if cfg.Output.Format != "sarif" {
		t.Errorf("Output.Format = %q, want sarif", cfg.Output.Format)
	}

	// Default for unspecified fields
	if !cfg.Nolint.Enabled {
		t.Error("Nolint.Enabled should default to true")
	}
	if !cfg.Analyzers.EnableAll {
		t.Error("Analyzers.EnableAll should default to true")
	}
}

func TestIsAnalyzerEnabled(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		analyzerName string
		want         bool
	}{
		{
			name:         "all enabled by default",
			yaml:         `version: 1`,
			analyzerName: "deferlint",
			want:         true,
		},
		{
			name: "specific analyzer disabled",
			yaml: `
version: 1
analyzers:
  disable:
    - deferlint
`,
			analyzerName: "deferlint",
			want:         false,
		},
		{
			name: "other analyzer still enabled when one disabled",
			yaml: `
version: 1
analyzers:
  disable:
    - deferlint
`,
			analyzerName: "errorlint",
			want:         true,
		},
		{
			name: "all disabled then enable specific",
			yaml: `
version: 1
analyzers:
  enable-all: false
  enable:
    - deferlint
`,
			analyzerName: "deferlint",
			want:         true,
		},
		{
			name: "all disabled, analyzer not in enable list",
			yaml: `
version: 1
analyzers:
  enable-all: false
  enable:
    - deferlint
`,
			analyzerName: "errorlint",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromReader([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("LoadFromReader() error = %v", err)
			}

			got := cfg.IsAnalyzerEnabled(tt.analyzerName)
			if got != tt.want {
				t.Errorf("IsAnalyzerEnabled(%q) = %v, want %v", tt.analyzerName, got, tt.want)
			}
		})
	}
}

func TestLoadWithOverridesExplicitPath(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	explicitPath := filepath.Join(tmpDir, "explicit-config.yml")
	configContent := `
version: 1
output:
  format: sarif
`
	if err := os.WriteFile(explicitPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load with explicit path should use that config
	cfg, err := config.LoadWithOverrides("/some/other/dir", explicitPath)
	if err != nil {
		t.Fatalf("LoadWithOverrides() error = %v", err)
	}
	if cfg.Output.Format != "sarif" {
		t.Errorf("Output.Format = %q, want sarif", cfg.Output.Format)
	}
}

func TestLoadWithOverridesExplicitPathNotFound(t *testing.T) {
	_, err := config.LoadWithOverrides("/some/dir", "/nonexistent/config.yml")
	if err == nil {
		t.Error("LoadWithOverrides() expected error for non-existent explicit path")
	}
}

func TestLoadWithOverridesFallsBackToDiscovery(t *testing.T) {
	// Create a temp directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".go-ai-lint.yml")
	configContent := `
version: 1
output:
  format: json
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load with empty explicit path should use discovery
	cfg, err := config.LoadWithOverrides(tmpDir, "")
	if err != nil {
		t.Fatalf("LoadWithOverrides() error = %v", err)
	}
	if cfg.Output.Format != formatJSON {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
	}
}

func TestToYAML(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		wantSubstr []string
	}{
		{
			name: "default config serializes to YAML",
			yaml: `version: 1`,
			wantSubstr: []string{
				"version: 1",
				"format: text",
				"enable-all: true",
			},
		},
		{
			name: "custom config preserves values",
			yaml: `
version: 1
output:
  format: json
severity:
  min-severity: high
`,
			wantSubstr: []string{
				"format: json",
				"min-severity: high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromReader([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("LoadFromReader() error = %v", err)
			}

			got, err := cfg.ToYAML()
			if err != nil {
				t.Fatalf("ToYAML() error = %v", err)
			}

			for _, substr := range tt.wantSubstr {
				if !contains(got, substr) {
					t.Errorf("ToYAML() output missing %q\nGot:\n%s", substr, got)
				}
			}
		})
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerateDefaultConfig(t *testing.T) {
	// When: Generate default config template
	content := config.GenerateDefaultConfig()

	// Then: Config contains version
	if !contains(content, "version: 1") {
		t.Error("GenerateDefaultConfig() missing 'version: 1'")
	}

	// Then: Config contains helpful comments
	if !contains(content, "# go-ai-lint configuration") {
		t.Error("GenerateDefaultConfig() missing header comment")
	}

	// Then: Config contains run section
	if !contains(content, "run:") {
		t.Error("GenerateDefaultConfig() missing 'run:' section")
	}

	// Then: Config contains output section
	if !contains(content, "output:") {
		t.Error("GenerateDefaultConfig() missing 'output:' section")
	}

	// Then: Config contains analyzers section
	if !contains(content, "analyzers:") {
		t.Error("GenerateDefaultConfig() missing 'analyzers:' section")
	}

	// Then: Config contains severity section
	if !contains(content, "severity:") {
		t.Error("GenerateDefaultConfig() missing 'severity:' section")
	}

	// Then: Config contains nolint section
	if !contains(content, "nolint:") {
		t.Error("GenerateDefaultConfig() missing 'nolint:' section")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "valid config",
			yaml:    `version: 1`,
			wantErr: false,
		},
		{
			name: "invalid output format",
			yaml: `
version: 1
output:
  format: invalid
`,
			wantErr: true,
		},
		{
			name: "invalid min-severity",
			yaml: `
version: 1
severity:
  min-severity: invalid
`,
			wantErr: true,
		},
		{
			name: "valid output formats",
			yaml: `
version: 1
output:
  format: json
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromReader([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("LoadFromReader() error = %v", err)
			}

			err = cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeEnableFlagAddsAnalyzer(t *testing.T) {
	// Given: Config has enable-all: false and no analyzers enabled
	yaml := `
version: 1
analyzers:
  enable-all: false
  enable: []
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// When: Merge with --enable=deferlint,errorlint
	overrides := config.CLIOverrides{
		Enable: []string{"deferlint", "errorlint"},
	}
	cfg.Merge(overrides)

	// Then: deferlint and errorlint are enabled, others disabled
	if !cfg.IsAnalyzerEnabled("deferlint") {
		t.Error("deferlint should be enabled after merge")
	}
	if !cfg.IsAnalyzerEnabled("errorlint") {
		t.Error("errorlint should be enabled after merge")
	}
	if cfg.IsAnalyzerEnabled("optionlint") {
		t.Error("optionlint should remain disabled (not in enable list)")
	}
}

func TestMergeDisableFlagRemovesAnalyzer(t *testing.T) {
	// Given: Config has enable-all: true
	yaml := `
version: 1
analyzers:
  enable-all: true
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// When: Merge with --disable=optionlint
	overrides := config.CLIOverrides{
		Disable: []string{"optionlint"},
	}
	cfg.Merge(overrides)

	// Then: optionlint is disabled, all others remain enabled
	if cfg.IsAnalyzerEnabled("optionlint") {
		t.Error("optionlint should be disabled after merge")
	}
	if !cfg.IsAnalyzerEnabled("deferlint") {
		t.Error("deferlint should remain enabled (enable-all: true)")
	}
	if !cfg.IsAnalyzerEnabled("errorlint") {
		t.Error("errorlint should remain enabled (enable-all: true)")
	}
}

func TestMergeMinSeverityFilters(t *testing.T) {
	// Given: Config has min-severity: low
	yaml := `
version: 1
severity:
  min-severity: low
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// When: Merge with --min-severity=high
	overrides := config.CLIOverrides{
		MinSeverity: "high",
	}
	cfg.Merge(overrides)

	// Then: min-severity is high
	if cfg.Severity.MinSeverity != "high" {
		t.Errorf("Severity.MinSeverity = %q, want high", cfg.Severity.MinSeverity)
	}
}

func TestMergeFormatFlagChangesOutput(t *testing.T) {
	// Given: Config has format: text
	yaml := `
version: 1
output:
  format: text
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// When: Merge with --format=json
	overrides := config.CLIOverrides{
		Format: formatJSON,
	}
	cfg.Merge(overrides)

	// Then: Output format is JSON
	if cfg.Output.Format != formatJSON {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
	}
}

func TestMergeCLIFlagsOverrideConfig(t *testing.T) {
	// Given: Config file has format: text, min-severity: low
	yaml := `
version: 1
output:
  format: text
severity:
  min-severity: low
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// When: Merge with --format=json --min-severity=high
	overrides := config.CLIOverrides{
		Format:      formatJSON,
		MinSeverity: "high",
	}
	cfg.Merge(overrides)

	// Then: Merged config uses CLI values
	if cfg.Output.Format != formatJSON {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatJSON)
	}
	if cfg.Severity.MinSeverity != "high" {
		t.Errorf("Severity.MinSeverity = %q, want high", cfg.Severity.MinSeverity)
	}
}

func TestMergeEmptyOverridesPreservesConfig(t *testing.T) {
	// Given: Config has specific values
	yaml := `
version: 1
output:
  format: sarif
severity:
  min-severity: medium
analyzers:
  enable-all: false
  enable:
    - deferlint
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	// When: Merge with empty overrides
	overrides := config.CLIOverrides{}
	cfg.Merge(overrides)

	// Then: Original config is preserved
	if cfg.Output.Format != formatSarif {
		t.Errorf("Output.Format = %q, want %s", cfg.Output.Format, formatSarif)
	}
	if cfg.Severity.MinSeverity != "medium" {
		t.Errorf("Severity.MinSeverity = %q, want medium", cfg.Severity.MinSeverity)
	}
	if !cfg.IsAnalyzerEnabled("deferlint") {
		t.Error("deferlint should remain enabled")
	}
}
