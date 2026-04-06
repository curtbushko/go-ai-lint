package errorlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/errorlint"
)

func TestErrorlint(t *testing.T) {
	errorLinter := errorlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := errorLinter.Analyzer()
	require.NotNil(t, analysisAnalyzer, "Analyzer() returned nil")
	assert.Equal(t, "errorlint", analysisAnalyzer.Name, "Analyzer name should be errorlint")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "errorlint")
}

func TestAnalyzerName(t *testing.T) {
	errorLinter := errorlint.New()
	assert.Equal(t, "errorlint", errorLinter.Name(), "Name() should return errorlint")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
