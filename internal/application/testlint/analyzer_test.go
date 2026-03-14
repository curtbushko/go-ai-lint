package testlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/testlint"
)

func TestTestlint_AIL130(t *testing.T) {
	testLinter := testlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := testLinter.Analyzer()
	if analysisAnalyzer.Name != "testlint" {
		t.Errorf("Analyzer name = %q, want testlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "testlint")
}

func TestAnalyzerName(t *testing.T) {
	analyzer := testlint.New()
	if analyzer.Name() != "testlint" {
		t.Errorf("Name() = %q, want testlint", analyzer.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	// Navigate from internal/application/testlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
