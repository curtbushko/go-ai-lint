package contextlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/contextlint"
)

func TestContextTODO(t *testing.T) {
	ctxLinter := contextlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := ctxLinter.Analyzer()
	if analysisAnalyzer.Name != "contextlint" {
		t.Errorf("Analyzer name = %q, want contextlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "contextlint")
}

func TestAnalyzerName(t *testing.T) {
	ctxLinter := contextlint.New()
	if ctxLinter.Name() != "contextlint" {
		t.Errorf("Name() = %q, want contextlint", ctxLinter.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	// Navigate from internal/core/analyzers/contextlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
