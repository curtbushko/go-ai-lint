package initlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/initlint"
)

func TestInitlint(t *testing.T) {
	initLinter := initlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := initLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "initlint" {
		t.Errorf("Analyzer name = %q, want initlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "initlint")
}

func TestAnalyzerName(t *testing.T) {
	initLinter := initlint.New()
	if initLinter.Name() != "initlint" {
		t.Errorf("Name() = %q, want initlint", initLinter.Name())
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
