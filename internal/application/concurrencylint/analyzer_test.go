package concurrencylint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/concurrencylint"
)

func TestConcurrencylint(t *testing.T) {
	concurrencyLinter := concurrencylint.New()

	// Verify analyzer metadata
	analysisAnalyzer := concurrencyLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "concurrencylint" {
		t.Errorf("Analyzer name = %q, want concurrencylint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "concurrencylint")
}

func TestAnalyzerName(t *testing.T) {
	concurrencyLinter := concurrencylint.New()
	if concurrencyLinter.Name() != "concurrencylint" {
		t.Errorf("Name() = %q, want concurrencylint", concurrencyLinter.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
