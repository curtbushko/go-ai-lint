package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/go-ai-lint/internal/config"
)

// Test constants to avoid goconst lint warnings.
const (
	severityHigh = "high"
	formatJSON   = "json"
)

// resetCLI resets viper and cobra state between tests.
func resetCLI() {
	viper.Reset()
	cfgFile = "" // Reset global config file path

	rootCmd.ResetFlags()
	rootCmd.ResetCommands()

	// Re-initialize flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().String("enable", "", "analyzers to enable")
	rootCmd.PersistentFlags().String("disable", "", "analyzers to disable")
	rootCmd.PersistentFlags().String("min-severity", "", "minimum severity")
	rootCmd.PersistentFlags().String("format", "", "output format")

	_ = viper.BindPFlag("enable", rootCmd.PersistentFlags().Lookup("enable"))
	_ = viper.BindPFlag("disable", rootCmd.PersistentFlags().Lookup("disable"))
	_ = viper.BindPFlag("min-severity", rootCmd.PersistentFlags().Lookup("min-severity"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

	// Reset init command flags
	initCmd.ResetFlags()
	initCmd.Flags().Bool("force", false, "overwrite existing config file")
	initCmd.Flags().String("dir", "", "directory for config file")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(showConfigCmd)
}

// executeCommand executes the root command with the given args and captures output.
func executeCommand(args ...string) (stdout string, err error) {
	resetCLI()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs(args)

	err = rootCmd.Execute()
	return stdoutBuf.String(), err
}

func TestInitCreatesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, config.ConfigFileName)

	stdout, err := executeCommand("init", "--dir="+tmpDir)

	require.NoError(t, err)
	assert.FileExists(t, configPath)
	assert.Contains(t, stdout, configPath)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "version: 1")
	assert.Contains(t, string(content), "# go-ai-lint configuration")
}

func TestInitRefusesOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, config.ConfigFileName)
	existingContent := "version: 1\noutput:\n  format: json\n"
	err := os.WriteFile(configPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	_, err = executeCommand("init", "--dir="+tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify file was not modified
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(content))
}

func TestInitForceOverwrites(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, config.ConfigFileName)
	existingContent := "version: 1\noutput:\n  format: json\n"
	err := os.WriteFile(configPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	_, err = executeCommand("init", "--force", "--dir="+tmpDir)

	require.NoError(t, err)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "format: json")
	assert.Contains(t, string(content), "# go-ai-lint configuration")
}

func TestShowConfigDisplaysYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yml")
	configContent := `
version: 1
output:
  format: sarif
severity:
  min-severity: critical
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	stdout, err := executeCommand("show-config", "--config="+configPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "format: sarif")
	assert.Contains(t, stdout, "min-severity: critical")
}

func TestShowConfigShowsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()

	stdout, err := executeCommand("show-config")

	require.NoError(t, err)
	assert.Contains(t, stdout, "format: text")
	assert.Contains(t, stdout, "# Source: defaults")
}

func TestShowConfigIncludesSource(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "my-config.yml")
	err := os.WriteFile(configPath, []byte("version: 1"), 0644)
	require.NoError(t, err)

	stdout, err := executeCommand("show-config", "--config="+configPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "# Source: "+configPath)
}

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single item",
			input:    "deferlint",
			expected: []string{"deferlint"},
		},
		{
			name:     "multiple items",
			input:    "deferlint,errorlint,optionlint",
			expected: []string{"deferlint", "errorlint", "optionlint"},
		},
		{
			name:     "with spaces",
			input:    "deferlint, errorlint , optionlint",
			expected: []string{"deferlint", "errorlint", "optionlint"},
		},
		{
			name:     "empty items ignored",
			input:    "deferlint,,errorlint",
			expected: []string{"deferlint", "errorlint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparated(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHelpCommand(t *testing.T) {
	stdout, err := executeCommand("--help")

	require.NoError(t, err)
	assert.Contains(t, stdout, "go-ai-lint")
	assert.Contains(t, stdout, "AI-generated Go code")
}

func TestRootCommandDescription(t *testing.T) {
	resetCLI()

	assert.Equal(t, "go-ai-lint [flags] [packages]", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "AI-generated Go code")
}

func TestInitCommandExists(t *testing.T) {
	resetCLI()

	var found bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "init" {
			found = true
			break
		}
	}
	assert.True(t, found, "init subcommand should exist")
}

func TestShowConfigCommandExists(t *testing.T) {
	resetCLI()

	var found bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "show-config" {
			found = true
			break
		}
	}
	assert.True(t, found, "show-config subcommand should exist")
}

func TestConfigFlagIsRegistered(t *testing.T) {
	resetCLI()

	flag := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, flag)
	assert.Equal(t, "config", flag.Name)
}

func TestEnableFlagIsRegistered(t *testing.T) {
	resetCLI()

	flag := rootCmd.PersistentFlags().Lookup("enable")
	require.NotNil(t, flag)
	assert.Equal(t, "enable", flag.Name)
}

func TestDisableFlagIsRegistered(t *testing.T) {
	resetCLI()

	flag := rootCmd.PersistentFlags().Lookup("disable")
	require.NotNil(t, flag)
	assert.Equal(t, "disable", flag.Name)
}

func TestMinSeverityFlagIsRegistered(t *testing.T) {
	resetCLI()

	flag := rootCmd.PersistentFlags().Lookup("min-severity")
	require.NotNil(t, flag)
	assert.Equal(t, "min-severity", flag.Name)
}

func TestFormatFlagIsRegistered(t *testing.T) {
	resetCLI()

	flag := rootCmd.PersistentFlags().Lookup("format")
	require.NotNil(t, flag)
	assert.Equal(t, "format", flag.Name)
}

func TestExecuteFunctionExists(t *testing.T) {
	// Execute function should be callable (we test it exists and is the right type)
	fn := Execute
	require.NotNil(t, fn)
}

func TestLoadConfigFromExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yml")
	configContent := `
version: 1
output:
  format: sarif
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfgFile = configPath
	defer func() { cfgFile = "" }()

	cfg, err := loadConfig()

	require.NoError(t, err)
	assert.Equal(t, "sarif", cfg.Output.Format)
}

func TestLoadConfigFromDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()

	cfgFile = ""

	cfg, err := loadConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	// Should have defaults
	assert.Equal(t, "text", cfg.Output.Format)
}

func TestApplyOverrides(t *testing.T) {
	viper.Reset()
	viper.Set("enable", "deferlint,errorlint")
	viper.Set("disable", "optionlint")
	viper.Set("min-severity", severityHigh)
	viper.Set("format", formatJSON)

	cfg := config.Default()
	applyOverrides(cfg)

	assert.Contains(t, cfg.Analyzers.Enable, "deferlint")
	assert.Contains(t, cfg.Analyzers.Enable, "errorlint")
	assert.Contains(t, cfg.Analyzers.Disable, "optionlint")
	assert.Equal(t, severityHigh, cfg.Severity.MinSeverity)
	assert.Equal(t, formatJSON, cfg.Output.Format)
}

func TestRunInitWithInvalidDir(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().String("dir", "/nonexistent/path/that/does/not/exist", "")

	err := runInit(cmd, nil)

	require.Error(t, err)
}
