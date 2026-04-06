package slicemaplint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/slicemaplint"
)

func TestNilMapWrite(t *testing.T) {
	mapLinter := slicemaplint.New()

	// Verify analyzer metadata
	analysisAnalyzer := mapLinter.Analyzer()
	assert.Equal(t, "slicemaplint", analysisAnalyzer.Name, "Analyzer name mismatch")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "slicemaplint")
}

func TestAnalyzerName(t *testing.T) {
	mapLinter := slicemaplint.New()
	assert.Equal(t, "slicemaplint", mapLinter.Name(), "Name() mismatch")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
