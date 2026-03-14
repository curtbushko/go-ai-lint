package cmdlint_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/curtbushko/go-ai-lint/internal/application/cmdlint"
)

func TestCmdlint_AIL120(t *testing.T) {
	cmdLinter := cmdlint.New()

	// Verify analyzer metadata
	analysisAnalyzer := cmdLinter.Analyzer()
	if analysisAnalyzer.Name != "cmdlint" {
		t.Errorf("Analyzer name = %q, want cmdlint", analysisAnalyzer.Name)
	}
	if analysisAnalyzer.Doc == "" {
		t.Error("Analyzer doc should not be empty")
	}

	// Run analysis on testdata/src/cmdlint which contains cmd/main.go test cases
	analysistest.Run(t, testdataDir(t), analysisAnalyzer, "cmdlint/cmd/...")
}

func TestAnalyzerName(t *testing.T) {
	cmdLinter := cmdlint.New()
	if cmdLinter.Name() != "cmdlint" {
		t.Errorf("Name() = %q, want cmdlint", cmdLinter.Name())
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get current test filename")
	}

	// Navigate from internal/application/cmdlint to project root, then to testdata
	return filepath.Join(filepath.Dir(testFilename), "..", "..", "..", "testdata")
}
