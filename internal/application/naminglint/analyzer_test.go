package naminglint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/naminglint"
)

func TestNaminglint(t *testing.T) {
	namingLinter := naminglint.New()

	// Verify analyzer metadata
	analysisAnalyzer := namingLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "naminglint" {
		t.Errorf("Analyzer name = %q, want naminglint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "naminglint")
}

func TestAnalyzerName(t *testing.T) {
	namingLinter := naminglint.New()
	if namingLinter.Name() != "naminglint" {
		t.Errorf("Name() = %q, want naminglint", namingLinter.Name())
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
