package stringlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/stringlint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestStringlint(t *testing.T) {
	stringLinter := stringlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := stringLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "stringlint" {
		t.Errorf("Analyzer name = %q, want stringlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "stringlint")
}

func TestAnalyzerName(t *testing.T) {
	stringLinter := stringlint.New()
	if stringLinter.Name() != "stringlint" {
		t.Errorf("Name() = %q, want stringlint", stringLinter.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "..", "testdata")
}
