package contextlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/contextlint"
)

func TestContextTODO(t *testing.T) {
	ctxLinter := contextlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := ctxLinter.Analyzer()
	assert.Equal(t, "contextlint", analysisAnalyzer.Name, "Analyzer name should be contextlint")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "contextlint")
}

func TestAnalyzerName(t *testing.T) {
	ctxLinter := contextlint.New()
	assert.Equal(t, "contextlint", ctxLinter.Name(), "Name() should return contextlint")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	// Navigate from internal/core/analyzers/contextlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
