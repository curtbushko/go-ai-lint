package optionlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/optionlint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOptionlint(t *testing.T) {
	optionLinter := optionlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := optionLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "optionlint" {
		t.Errorf("Analyzer name = %q, want optionlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "optionlint")
}

func TestAnalyzerName(t *testing.T) {
	optionLinter := optionlint.New()
	if optionLinter.Name() != "optionlint" {
		t.Errorf("Name() = %q, want optionlint", optionLinter.Name())
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
