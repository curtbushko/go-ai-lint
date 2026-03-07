package slicemaplint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/slicemaplint"
)

func TestNilMapWrite(t *testing.T) {
	mapLinter := slicemaplint.New()

	// Verify analyzer metadata
	analysisAnalyzer := mapLinter.Analyzer()
	if analysisAnalyzer.Name != "slicemaplint" {
		t.Errorf("Analyzer name = %q, want slicemaplint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "slicemaplint")
}

func TestAnalyzerName(t *testing.T) {
	mapLinter := slicemaplint.New()
	if mapLinter.Name() != "slicemaplint" {
		t.Errorf("Name() = %q, want slicemaplint", mapLinter.Name())
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
