package deferlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/deferlint"
)

func TestDeferInLoop(t *testing.T) {
	deferLinter := deferlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := deferLinter.Analyzer()
	assert.Equal(t, "deferlint", analysisAnalyzer.Name, "Analyzer name should be deferlint")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "deferlint")
}

func TestAnalyzerName(t *testing.T) {
	analyzer := deferlint.New()
	assert.Equal(t, "deferlint", analyzer.Name(), "Name() should return deferlint")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	// Navigate from internal/core/analyzers/deferlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
