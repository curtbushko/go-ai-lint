// Package main provides the CLI for go-ai-lint using cobra and viper.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/analysis"

	"github.com/curtbushko/go-ai-lint/internal/application/cmdlint"
	"github.com/curtbushko/go-ai-lint/internal/application/concurrencylint"
	"github.com/curtbushko/go-ai-lint/internal/application/contextlint"
	"github.com/curtbushko/go-ai-lint/internal/application/deferlint"
	"github.com/curtbushko/go-ai-lint/internal/application/errorlint"
	"github.com/curtbushko/go-ai-lint/internal/application/goroutinelint"
	"github.com/curtbushko/go-ai-lint/internal/application/initlint"
	"github.com/curtbushko/go-ai-lint/internal/application/interfacelint"
	"github.com/curtbushko/go-ai-lint/internal/application/iolint"
	"github.com/curtbushko/go-ai-lint/internal/application/naminglint"
	"github.com/curtbushko/go-ai-lint/internal/application/optionlint"
	"github.com/curtbushko/go-ai-lint/internal/application/paniclint"
	"github.com/curtbushko/go-ai-lint/internal/application/slicemaplint"
	"github.com/curtbushko/go-ai-lint/internal/application/stringlint"
	"github.com/curtbushko/go-ai-lint/internal/application/testlint"
	"github.com/curtbushko/go-ai-lint/internal/config"
	"github.com/curtbushko/go-ai-lint/internal/domain"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "go-ai-lint [flags] [packages]",
		Short: "A linter for detecting common mistakes in AI-generated Go code",
		Long: `go-ai-lint is a static analysis tool that detects common mistakes
in AI-generated Go code. It checks for issues with error handling,
concurrency, context usage, and more.

Examples:
  go-ai-lint ./...
  go-ai-lint --disable=cmdlint,testlint ./...
  go-ai-lint --min-severity=high ./pkg/...`,
		Args:         cobra.ArbitraryArgs,
		RunE:         runLint,
		SilenceUsage: true,
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Generate a default configuration file",
		Long:  `Generate a default .go-ai-lint.yml configuration file in the current directory.`,
		RunE:  runInit,
	}

	showConfigCmd = &cobra.Command{
		Use:   "show-config",
		Short: "Display the resolved configuration",
		Long:  `Display the fully resolved configuration including defaults and overrides.`,
		RunE:  runShowConfig,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .go-ai-lint.yml)")
	rootCmd.PersistentFlags().String("enable", "", "comma-separated list of analyzers to enable")
	rootCmd.PersistentFlags().String("disable", "", "comma-separated list of analyzers to disable")
	rootCmd.PersistentFlags().String("min-severity", "", "minimum severity to report: low, medium, high, critical")
	rootCmd.PersistentFlags().String("format", "", "output format: text, json, ai, sarif")

	// Bind flags to viper
	_ = viper.BindPFlag("enable", rootCmd.PersistentFlags().Lookup("enable"))
	_ = viper.BindPFlag("disable", rootCmd.PersistentFlags().Lookup("disable"))
	_ = viper.BindPFlag("min-severity", rootCmd.PersistentFlags().Lookup("min-severity"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

	// Init command flags
	initCmd.Flags().Bool("force", false, "overwrite existing config file")
	initCmd.Flags().String("dir", "", "directory for config file (defaults to current directory)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(showConfigCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cwd, err := os.Getwd()
		if err == nil {
			viper.AddConfigPath(cwd)
		}
		viper.SetConfigName(".go-ai-lint")
		viper.SetConfigType("yml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GO_AI_LINT")

	// Read config file if it exists (ignore errors for missing file)
	_ = viper.ReadInConfig()
}

func runLint(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Apply CLI overrides
	applyOverrides(cfg)

	// Apply nolint config setting
	domain.SetNolintEnabled(cfg.Nolint.Enabled)

	// Build list of all analyzers with their names
	allAnalyzers := map[string]*analysis.Analyzer{
		"cmdlint":         cmdlint.New().Analyzer(),
		"concurrencylint": concurrencylint.New().Analyzer(),
		"contextlint":     contextlint.New().Analyzer(),
		"deferlint":       deferlint.New().Analyzer(),
		"errorlint":       errorlint.New().Analyzer(),
		"goroutinelint":   goroutinelint.New().Analyzer(),
		"initlint":        initlint.New().Analyzer(),
		"interfacelint":   interfacelint.New().Analyzer(),
		"iolint":          iolint.New().Analyzer(),
		"naminglint":      naminglint.New().Analyzer(),
		"optionlint":      optionlint.New().Analyzer(),
		"paniclint":       paniclint.New().Analyzer(),
		"slicemaplint":    slicemaplint.New().Analyzer(),
		"stringlint":      stringlint.New().Analyzer(),
		"testlint":        testlint.New().Analyzer(),
	}

	// Filter analyzers based on config
	var analyzers []*analysis.Analyzer
	for name, a := range allAnalyzers {
		if cfg.IsAnalyzerEnabled(name) {
			analyzers = append(analyzers, a)
		}
	}

	// Default to current directory if no args
	packages := args
	if len(packages) == 0 {
		packages = []string{"."}
	}

	exitCode := RunAnalyzers(os.Stdout, os.Stderr, analyzers, packages)
	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	dir, _ := cmd.Flags().GetString("dir")

	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		dir = cwd
	}

	configPath := filepath.Join(dir, config.ConfigFileName)

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		if !force {
			return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
		}
	}

	// Generate and write default config
	content := config.GenerateDefaultConfig()
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", configPath)
	return nil
}

func runShowConfig(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	applyOverrides(cfg)

	yamlStr, err := cfg.ToYAML()
	if err != nil {
		return fmt.Errorf("serialize config: %w", err)
	}

	// Determine config source
	configSource := cfgFile
	if configSource == "" {
		configSource = viper.ConfigFileUsed()
	}
	if configSource == "" {
		cwd, _ := os.Getwd()
		configSource = config.FindConfigFile(cwd)
	}
	if configSource == "" {
		configSource = "defaults"
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "# Source: %s\n%s", configSource, yamlStr)
	return nil
}

func loadConfig() (*config.Config, error) {
	if cfgFile != "" {
		return config.LoadFromPath(cfgFile)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	return config.Load(cwd)
}

func applyOverrides(cfg *config.Config) {
	overrides := config.CLIOverrides{}

	if enable := viper.GetString("enable"); enable != "" {
		overrides.Enable = parseCommaSeparated(enable)
	}

	if disable := viper.GetString("disable"); disable != "" {
		overrides.Disable = parseCommaSeparated(disable)
	}

	if minSeverity := viper.GetString("min-severity"); minSeverity != "" {
		overrides.MinSeverity = minSeverity
	}

	if format := viper.GetString("format"); format != "" {
		overrides.Format = format
	}

	cfg.Merge(overrides)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
