package paniclint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/paniclint"
)

func TestPaniclint(t *testing.T) {
	panicLinter := paniclint.New()

	// Verify analyzer metadata
	analysisAnalyzer := panicLinter.Analyzer()
	require.NotNil(t, analysisAnalyzer, "Analyzer() returned nil")
	assert.Equal(t, "paniclint", analysisAnalyzer.Name, "Analyzer name mismatch")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "paniclint")
}

func TestAnalyzerName(t *testing.T) {
	panicLinter := paniclint.New()
	assert.Equal(t, "paniclint", panicLinter.Name(), "Name() mismatch")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
