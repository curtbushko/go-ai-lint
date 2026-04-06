package iolint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/iolint"
)

func TestIolintAnalyzer(t *testing.T) {
	ioLinter := iolint.New()

	// Verify analyzer metadata
	analysisAnalyzer := ioLinter.Analyzer()
	assert.Equal(t, "iolint", analysisAnalyzer.Name, "Analyzer name mismatch")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "iolint")
}

func TestAnalyzerName(t *testing.T) {
	analyzer := iolint.New()
	assert.Equal(t, "iolint", analyzer.Name(), "Name() mismatch")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	// Navigate from internal/application/iolint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
