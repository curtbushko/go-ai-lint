// Package linters provides the go-ai-lint plugin implementation.
package linters

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/curtbushko/go-ai-lint/internal/application/concurrencylint"
	"github.com/curtbushko/go-ai-lint/internal/application/contextlint"
	"github.com/curtbushko/go-ai-lint/internal/application/deferlint"
	"github.com/curtbushko/go-ai-lint/internal/application/errorlint"
	"github.com/curtbushko/go-ai-lint/internal/application/goroutinelint"
	"github.com/curtbushko/go-ai-lint/internal/application/initlint"
	"github.com/curtbushko/go-ai-lint/internal/application/interfacelint"
	"github.com/curtbushko/go-ai-lint/internal/application/naminglint"
	"github.com/curtbushko/go-ai-lint/internal/application/optionlint"
	"github.com/curtbushko/go-ai-lint/internal/application/paniclint"
	"github.com/curtbushko/go-ai-lint/internal/application/slicemaplint"
	"github.com/curtbushko/go-ai-lint/internal/application/stringlint"
)

func init() {
	register.Plugin("go-ai-lint", New)
}

// Settings defines the configuration structure for go-ai-lint.
type Settings struct {
	// EnabledAnalyzers specifies which analyzers to enable.
	// If empty, all analyzers are enabled.
	EnabledAnalyzers []string `json:"enabled_analyzers"`
}

// Plugin implements the go-ai-lint plugin for golangci-lint.
type Plugin struct {
	settings Settings
}

// New creates a new instance of the go-ai-lint plugin.
func New(settings any) (register.LinterPlugin, error) {
	pluginSettings, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	return &Plugin{settings: pluginSettings}, nil
}

// BuildAnalyzers returns the analyzers provided by this plugin.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	// Create all analyzers
	analyzers := []*analysis.Analyzer{
		deferlint.New().Analyzer(),
		contextlint.New().Analyzer(),
		slicemaplint.New().Analyzer(),
		goroutinelint.New().Analyzer(),
		errorlint.New().Analyzer(),
		concurrencylint.New().Analyzer(),
		naminglint.New().Analyzer(),
		interfacelint.New().Analyzer(),
		stringlint.New().Analyzer(),
		paniclint.New().Analyzer(),
		initlint.New().Analyzer(),
		optionlint.New().Analyzer(),
	}

	// Filter if specific analyzers are enabled
	if len(p.settings.EnabledAnalyzers) > 0 {
		enabledSet := make(map[string]bool)
		for _, name := range p.settings.EnabledAnalyzers {
			enabledSet[name] = true
		}

		filtered := make([]*analysis.Analyzer, 0, len(analyzers))
		for _, analyzer := range analyzers {
			if enabledSet[analyzer.Name] {
				filtered = append(filtered, analyzer)
			}
		}
		return filtered, nil
	}

	return analyzers, nil
}

// GetLoadMode returns the load mode required by this plugin.
//
//nolint:naminglint // AIL050: Required by golangci-lint plugin interface
func (p *Plugin) GetLoadMode() string {
	// Use LoadModeTypesInfo since analyzers need type information
	// (e.g., deferlint checks if methods return errors)
	return register.LoadModeTypesInfo
}
