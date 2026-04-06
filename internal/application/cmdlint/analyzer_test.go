package cmdlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/cmdlint"
)

func TestCmdlint_AIL120(t *testing.T) {
	cmdLinter := cmdlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := cmdLinter.Analyzer()
	assert.Equal(t, "cmdlint", analysisAnalyzer.Name, "Analyzer name should be cmdlint")
	assert.NotEmpty(t, analysisAnalyzer.Doc, "Analyzer doc should not be empty")

	// Run analysis on testdata/src/cmdlint which contains cmd/main.go test cases
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "cmdlint/cmd/...")
}

func TestAnalyzerName(t *testing.T) {
	cmdLinter := cmdlint.New()
	assert.Equal(t, "cmdlint", cmdLinter.Name(), "Name() should return cmdlint")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok, "unable to get current test filename")

	// Navigate from internal/application/cmdlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
