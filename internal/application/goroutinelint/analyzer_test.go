package goroutinelint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/goroutinelint"
)

func TestGoroutinelint(t *testing.T) {
	goroutineLinter := goroutinelint.New()

	// Verify analyzer metadata
	analysisAnalyzer := goroutineLinter.Analyzer()
	if analysisAnalyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analysisAnalyzer.Name != "goroutinelint" {
		t.Errorf("Analyzer name = %q, want goroutinelint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "goroutinelint")
}

func TestAnalyzerName(t *testing.T) {
	goroutineLinter := goroutinelint.New()
	if goroutineLinter.Name() != "goroutinelint" {
		t.Errorf("Name() = %q, want goroutinelint", goroutineLinter.Name())
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
