package iolint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/iolint"
)

func TestIolintAnalyzer(t *testing.T) {
	ioLinter := iolint.New()

	// Verify analyzer metadata
	analysisAnalyzer := ioLinter.Analyzer()
	if analysisAnalyzer.Name != "iolint" {
		t.Errorf("Analyzer name = %q, want iolint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "iolint")
}

func TestAnalyzerName(t *testing.T) {
	analyzer := iolint.New()
	if analyzer.Name() != "iolint" {
		t.Errorf("Name() = %q, want iolint", analyzer.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	// Navigate from internal/application/iolint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
