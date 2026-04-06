package testlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/testlint"
)

func TestTestlint_AIL130(t *testing.T) {
	testLinter := testlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := testLinter.Analyzer()
	assert.Equal(t, "testlint", analysisAnalyzer.Name, "Analyzer name mismatch")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "testlint")
}

func TestAnalyzerName(t *testing.T) {
	analyzer := testlint.New()
	assert.Equal(t, "testlint", analyzer.Name(), "Name() mismatch")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	// Navigate from internal/application/testlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
