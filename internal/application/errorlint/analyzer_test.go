package errorlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/errorlint"
)

func TestErrorlint(t *testing.T) {
	errorLinter := errorlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := errorLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "errorlint" {
		t.Errorf("Analyzer name = %q, want errorlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "errorlint")
}

func TestAnalyzerName(t *testing.T) {
	errorLinter := errorlint.New()
	if errorLinter.Name() != "errorlint" {
		t.Errorf("Name() = %q, want errorlint", errorLinter.Name())
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
