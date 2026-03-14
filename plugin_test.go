package linters

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestGoAILintPlugin(t *testing.T) {
	newPlugin, err := register.GetPlugin("go-ai-lint")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)
	require.NotEmpty(t, analyzers, "should have at least one analyzer")

	// Test deferlint analyzer
	var deferlintAnalyzer *struct {
		found bool
		index int
	}
	for idx, analyzer := range analyzers {
		if analyzer.Name == "deferlint" {
			deferlintAnalyzer = &struct {
				found bool
				index int
			}{found: true, index: idx}
			break
		}
	}
	require.NotNil(t, deferlintAnalyzer, "deferlint analyzer should be present")

	// Run deferlint on testdata
	analysistest.Run(t, testdataDir(t), analyzers[deferlintAnalyzer.index], "deferlint")
}

func TestGoAILintPluginWithSettings(t *testing.T) {
	newPlugin, err := register.GetPlugin("go-ai-lint")
	require.NoError(t, err)

	// Test with specific analyzers enabled
	settings := map[string]any{
		"enabled_analyzers": []string{"deferlint"},
	}

	plugin, err := newPlugin(settings)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)
	require.Len(t, analyzers, 1, "should have exactly one analyzer when filtered")
	require.Equal(t, "deferlint", analyzers[0].Name)
}

func TestGoAILintPluginLoadMode(t *testing.T) {
	newPlugin, err := register.GetPlugin("go-ai-lint")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	loadMode := plugin.GetLoadMode()
	require.NotEmpty(t, loadMode, "load mode should not be empty")
}

func TestGoAILintPluginAllAnalyzers(t *testing.T) {
	// Verify all 12 analyzers are registered
	expectedAnalyzers := []string{
		"cmdlint",
		"concurrencylint",
		"contextlint",
		"deferlint",
		"errorlint",
		"goroutinelint",
		"initlint",
		"interfacelint",
		"naminglint",
		"optionlint",
		"paniclint",
		"slicemaplint",
		"stringlint",
	}

	newPlugin, err := register.GetPlugin("go-ai-lint")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)
	require.Len(t, analyzers, len(expectedAnalyzers), "should have exactly %d analyzers", len(expectedAnalyzers))

	// Build a map of analyzer names
	analyzerNames := make(map[string]bool)
	for _, a := range analyzers {
		analyzerNames[a.Name] = true
	}

	// Verify each expected analyzer is present
	for _, name := range expectedAnalyzers {
		require.True(t, analyzerNames[name], "analyzer %s should be registered", name)
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		require.Fail(t, "unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
