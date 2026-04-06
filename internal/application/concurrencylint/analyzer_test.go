package concurrencylint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/concurrencylint"
)

func TestConcurrencylint(t *testing.T) {
	concurrencyLinter := concurrencylint.New()

	// Verify analyzer metadata
	analysisAnalyzer := concurrencyLinter.Analyzer()
	require.NotNil(t, analysisAnalyzer, "Analyzer() returned nil")
	assert.Equal(t, "concurrencylint", analysisAnalyzer.Name, "Analyzer name should be concurrencylint")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "concurrencylint")
}

func TestAnalyzerName(t *testing.T) {
	concurrencyLinter := concurrencylint.New()
	assert.Equal(t, "concurrencylint", concurrencyLinter.Name(), "Name() should return concurrencylint")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
