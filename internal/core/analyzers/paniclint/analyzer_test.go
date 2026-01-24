package paniclint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/analyzers/paniclint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPaniclint(t *testing.T) {
	panicLinter := paniclint.New()

	// Verify analyzer metadata
	analysisAnalyzer := panicLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "paniclint" {
		t.Errorf("Analyzer name = %q, want paniclint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "paniclint")
}

func TestAnalyzerName(t *testing.T) {
	panicLinter := paniclint.New()
	if panicLinter.Name() != "paniclint" {
		t.Errorf("Name() = %q, want paniclint", panicLinter.Name())
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
