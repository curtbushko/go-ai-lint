package deferlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/deferlint"
)

func TestDeferInLoop(t *testing.T) {
	deferLinter := deferlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := deferLinter.Analyzer()
	if analysisAnalyzer.Name != "deferlint" {
		t.Errorf("Analyzer name = %q, want deferlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "deferlint")
}

func TestAnalyzerName(t *testing.T) {
	analyzer := deferlint.New()
	if analyzer.Name() != "deferlint" {
		t.Errorf("Name() = %q, want deferlint", analyzer.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	// Navigate from internal/core/analyzers/deferlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
