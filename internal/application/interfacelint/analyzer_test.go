package interfacelint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/interfacelint"
)

func TestInterfacelint(t *testing.T) {
	interfaceLinter := interfacelint.New()

	// Verify analyzer metadata
	analysisAnalyzer := interfaceLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "interfacelint" {
		t.Errorf("Analyzer name = %q, want interfacelint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "interfacelint")
}

func TestAnalyzerName(t *testing.T) {
	interfaceLinter := interfacelint.New()
	if interfaceLinter.Name() != "interfacelint" {
		t.Errorf("Name() = %q, want interfacelint", interfaceLinter.Name())
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
