package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.Equal(t, 1, cfg.Version, "Default().Version")

	// Check run defaults
	assert.Equal(t, 5*time.Minute, cfg.Run.Timeout, "Default().Run.Timeout")
	assert.Equal(t, 0, cfg.Run.Concurrency, "Default().Run.Concurrency should be 0 (auto)")

	// Check output defaults
	assert.Equal(t, "text", cfg.Output.Format, "Default().Output.Format")
	assert.True(t, cfg.Output.PrintAnalyzerName, "Default().Output.PrintAnalyzerName should be true")
	assert.Equal(t, "file", cfg.Output.SortBy, "Default().Output.SortBy")

	// Check nolint defaults
	assert.True(t, cfg.Nolint.Enabled, "Default().Nolint.Enabled should be true")
	assert.False(t, cfg.Nolint.RequireSpecific, "Default().Nolint.RequireSpecific should be false")

	// Check analyzer defaults
	assert.True(t, cfg.Analyzers.EnableAll, "Default().Analyzers.EnableAll should be true")

	// Check severity defaults
	assert.Equal(t, "low", cfg.Severity.MinSeverity, "Default().Severity.MinSeverity")
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
				assert.Equal(t, 1, cfg.Version)
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
				assert.Equal(t, 10*time.Minute, cfg.Run.Timeout)
				assert.Equal(t, 4, cfg.Run.Concurrency)
				assert.Len(t, cfg.Run.SkipDirs, 2)
				assert.Len(t, cfg.Run.SkipFiles, 1)
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
				assert.Equal(t, formatJSON, cfg.Output.Format)
				assert.False(t, cfg.Output.PrintAnalyzerName, "PrintAnalyzerName should be false")
				assert.Equal(t, "severity", cfg.Output.SortBy)
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
				assert.False(t, cfg.Nolint.Enabled, "Nolint.Enabled should be false")
				assert.True(t, cfg.Nolint.RequireSpecific, "Nolint.RequireSpecific should be true")
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
				assert.False(t, cfg.Analyzers.EnableAll, "Analyzers.EnableAll should be false")
				assert.Len(t, cfg.Analyzers.Disable, 2)
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
				assert.Equal(t, "medium", cfg.Severity.MinSeverity)
				assert.Len(t, cfg.Severity.ErrorOn, 2)
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
			if tt.wantErr {
				assert.Error(t, err, "LoadFromReader() should return error")
				return
			}
			require.NoError(t, err, "LoadFromReader() should not return error")
			if tt.check != nil {
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
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to write config file")

	// Load config starting from tmpDir
	cfg, err := config.Load(tmpDir)
	require.NoError(t, err, "Load() should not return error")

	assert.Equal(t, formatJSON, cfg.Output.Format)
}

func TestLoadConfigFromParentDir(t *testing.T) {
	// Create a temp directory structure: parent/.go-ai-lint.yml and parent/child/
	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "parent")
	childDir := filepath.Join(parentDir, "child")

	err := os.MkdirAll(childDir, 0755)
	require.NoError(t, err, "failed to create directories")

	configPath := filepath.Join(parentDir, ".go-ai-lint.yml")
	configContent := `
version: 1
output:
  format: ai
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to write config file")

	// Load config starting from childDir (should find in parentDir)
	cfg, err := config.Load(childDir)
	require.NoError(t, err, "Load() should not return error")

	assert.Equal(t, "ai", cfg.Output.Format)
}

func TestLoadConfigReturnsDefaultsWhenNoConfigFound(t *testing.T) {
	// Create a temp directory with no config file
	tmpDir := t.TempDir()

	cfg, err := config.Load(tmpDir)
	require.NoError(t, err, "Load() should not return error")

	// Should return defaults
	defaultCfg := config.Default()
	assert.Equal(t, defaultCfg.Version, cfg.Version)
	assert.Equal(t, defaultCfg.Output.Format, cfg.Output.Format)
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
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to write config file")

	cfg, err := config.LoadFromPath(configPath)
	require.NoError(t, err, "LoadFromPath() should not return error")

	assert.Equal(t, "critical", cfg.Severity.MinSeverity)
}

func TestLoadConfigFromExplicitPathNotFound(t *testing.T) {
	_, err := config.LoadFromPath("/nonexistent/path/config.yml")
	assert.Error(t, err, "LoadFromPath() should return error for non-existent file")
}

func TestConfigMergeWithDefaults(t *testing.T) {
	// Load a partial config - should have defaults for unspecified fields
	yaml := `
version: 1
output:
  format: sarif
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	require.NoError(t, err, "LoadFromReader() should not return error")

	// Explicitly set field
	assert.Equal(t, "sarif", cfg.Output.Format)

	// Default for unspecified fields
	assert.True(t, cfg.Nolint.Enabled, "Nolint.Enabled should default to true")
	assert.True(t, cfg.Analyzers.EnableAll, "Analyzers.EnableAll should default to true")
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
			require.NoError(t, err, "LoadFromReader() should not return error")

			got := cfg.IsAnalyzerEnabled(tt.analyzerName)
			assert.Equal(t, tt.want, got, "IsAnalyzerEnabled(%q)", tt.analyzerName)
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
	err := os.WriteFile(explicitPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to write config file")

	// Load with explicit path should use that config
	cfg, err := config.LoadWithOverrides("/some/other/dir", explicitPath)
	require.NoError(t, err, "LoadWithOverrides() should not return error")
	assert.Equal(t, "sarif", cfg.Output.Format)
}

func TestLoadWithOverridesExplicitPathNotFound(t *testing.T) {
	_, err := config.LoadWithOverrides("/some/dir", "/nonexistent/config.yml")
	assert.Error(t, err, "LoadWithOverrides() should return error for non-existent explicit path")
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
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to write config file")

	// Load with empty explicit path should use discovery
	cfg, err := config.LoadWithOverrides(tmpDir, "")
	require.NoError(t, err, "LoadWithOverrides() should not return error")
	assert.Equal(t, formatJSON, cfg.Output.Format)
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
			require.NoError(t, err, "LoadFromReader() should not return error")

			got, err := cfg.ToYAML()
			require.NoError(t, err, "ToYAML() should not return error")

			for _, substr := range tt.wantSubstr {
				assert.Contains(t, got, substr, "ToYAML() output should contain %q", substr)
			}
		})
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	// When: Generate default config template
	content := config.GenerateDefaultConfig()

	// Then: Config contains version
	assert.Contains(t, content, "version: 1", "GenerateDefaultConfig() should contain 'version: 1'")

	// Then: Config contains helpful comments
	assert.Contains(t, content, "# go-ai-lint configuration", "GenerateDefaultConfig() should contain header comment")

	// Then: Config contains run section
	assert.Contains(t, content, "run:", "GenerateDefaultConfig() should contain 'run:' section")

	// Then: Config contains output section
	assert.Contains(t, content, "output:", "GenerateDefaultConfig() should contain 'output:' section")

	// Then: Config contains analyzers section
	assert.Contains(t, content, "analyzers:", "GenerateDefaultConfig() should contain 'analyzers:' section")

	// Then: Config contains severity section
	assert.Contains(t, content, "severity:", "GenerateDefaultConfig() should contain 'severity:' section")

	// Then: Config contains nolint section
	assert.Contains(t, content, "nolint:", "GenerateDefaultConfig() should contain 'nolint:' section")
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
			require.NoError(t, err, "LoadFromReader() should not return error")

			err = cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Validate() should return error")
			} else {
				assert.NoError(t, err, "Validate() should not return error")
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
	require.NoError(t, err, "LoadFromReader() should not return error")

	// When: Merge with --enable=deferlint,errorlint
	overrides := config.CLIOverrides{
		Enable: []string{"deferlint", "errorlint"},
	}
	cfg.Merge(overrides)

	// Then: deferlint and errorlint are enabled, others disabled
	assert.True(t, cfg.IsAnalyzerEnabled("deferlint"), "deferlint should be enabled after merge")
	assert.True(t, cfg.IsAnalyzerEnabled("errorlint"), "errorlint should be enabled after merge")
	assert.False(t, cfg.IsAnalyzerEnabled("optionlint"), "optionlint should remain disabled (not in enable list)")
}

func TestMergeDisableFlagRemovesAnalyzer(t *testing.T) {
	// Given: Config has enable-all: true
	yaml := `
version: 1
analyzers:
  enable-all: true
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	require.NoError(t, err, "LoadFromReader() should not return error")

	// When: Merge with --disable=optionlint
	overrides := config.CLIOverrides{
		Disable: []string{"optionlint"},
	}
	cfg.Merge(overrides)

	// Then: optionlint is disabled, all others remain enabled
	assert.False(t, cfg.IsAnalyzerEnabled("optionlint"), "optionlint should be disabled after merge")
	assert.True(t, cfg.IsAnalyzerEnabled("deferlint"), "deferlint should remain enabled (enable-all: true)")
	assert.True(t, cfg.IsAnalyzerEnabled("errorlint"), "errorlint should remain enabled (enable-all: true)")
}

func TestMergeMinSeverityFilters(t *testing.T) {
	// Given: Config has min-severity: low
	yaml := `
version: 1
severity:
  min-severity: low
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	require.NoError(t, err, "LoadFromReader() should not return error")

	// When: Merge with --min-severity=high
	overrides := config.CLIOverrides{
		MinSeverity: "high",
	}
	cfg.Merge(overrides)

	// Then: min-severity is high
	assert.Equal(t, "high", cfg.Severity.MinSeverity)
}

func TestMergeFormatFlagChangesOutput(t *testing.T) {
	// Given: Config has format: text
	yaml := `
version: 1
output:
  format: text
`
	cfg, err := config.LoadFromReader([]byte(yaml))
	require.NoError(t, err, "LoadFromReader() should not return error")

	// When: Merge with --format=json
	overrides := config.CLIOverrides{
		Format: formatJSON,
	}
	cfg.Merge(overrides)

	// Then: Output format is JSON
	assert.Equal(t, formatJSON, cfg.Output.Format)
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
	require.NoError(t, err, "LoadFromReader() should not return error")

	// When: Merge with --format=json --min-severity=high
	overrides := config.CLIOverrides{
		Format:      formatJSON,
		MinSeverity: "high",
	}
	cfg.Merge(overrides)

	// Then: Merged config uses CLI values
	assert.Equal(t, formatJSON, cfg.Output.Format)
	assert.Equal(t, "high", cfg.Severity.MinSeverity)
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
	require.NoError(t, err, "LoadFromReader() should not return error")

	// When: Merge with empty overrides
	overrides := config.CLIOverrides{}
	cfg.Merge(overrides)

	// Then: Original config is preserved
	assert.Equal(t, formatSarif, cfg.Output.Format)
	assert.Equal(t, "medium", cfg.Severity.MinSeverity)
	assert.True(t, cfg.IsAnalyzerEnabled("deferlint"), "deferlint should remain enabled")
}

func TestExampleConfigParseable(t *testing.T) {
	// Given: The example config file exists in the repo root
	// This test validates that .go-ai-lint.yml.example is valid YAML
	// and can be parsed by the config loader.

	// Find the example config relative to this test file
	// Walk up to find the repo root where .go-ai-lint.yml.example lives
	exampleConfigPath := findExampleConfig(t)

	// When: Load the example config
	cfg, err := config.LoadFromPath(exampleConfigPath)

	// Then: No parse errors
	require.NoError(t, err, "LoadFromPath(%s) should not return error", exampleConfigPath)

	// Then: Config validates successfully
	err = cfg.Validate()
	require.NoError(t, err, "Validate() should not return error")

	// Then: Config has expected structure
	assert.Equal(t, 1, cfg.Version)

	// Then: Output format is valid
	assert.NotEmpty(t, cfg.Output.Format, "Output.Format should not be empty")

	// Then: Severity is valid
	assert.NotEmpty(t, cfg.Severity.MinSeverity, "Severity.MinSeverity should not be empty")
}

// findExampleConfig walks up directories to find .go-ai-lint.yml.example.
func findExampleConfig(t *testing.T) string {
	t.Helper()

	// Get the directory of this test file
	// Start from the current working directory during test execution
	cwd, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	// Walk up to find the example config
	dir := cwd
	for {
		examplePath := filepath.Join(dir, ".go-ai-lint.yml.example")
		if _, err := os.Stat(examplePath); err == nil {
			return examplePath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			require.Fail(t, "could not find .go-ai-lint.yml.example starting from %s", cwd)
		}
		dir = parent
	}
}
