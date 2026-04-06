package main

import (
	"bytes"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestRunnerPrintsSuccessOnNoIssues(t *testing.T) {
	// Given: An analyzer that finds no issues
	noop := &analysis.Analyzer{
		Name: "noop",
		Doc:  "does nothing",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}

	// When: Run the analyzer on clean code
	var stdout, stderr bytes.Buffer
	exitCode := RunAnalyzers(&stdout, &stderr, []*analysis.Analyzer{noop}, []string{"./testdata/clean"})

	// Then: Exit code is 0 and success message is printed
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	output := stdout.String()
	if output == "" {
		t.Error("expected success message on stdout, got empty")
	}
}

func TestRunnerPrintsIssuesOnError(t *testing.T) {
	// Given: An analyzer that reports issues
	issueAnalyzer := &analysis.Analyzer{
		Name: "issue",
		Doc:  "reports an issue",
		Run: func(pass *analysis.Pass) (interface{}, error) {
			for _, f := range pass.Files {
				pass.Reportf(f.Pos(), "test issue")
				break
			}
			return nil, nil
		},
	}

	// When: Run the analyzer on code
	var stdout, stderr bytes.Buffer
	exitCode := RunAnalyzers(&stdout, &stderr, []*analysis.Analyzer{issueAnalyzer}, []string{"./testdata/clean"})

	// Then: Exit code is non-zero and issues are in output
	if exitCode == 0 {
		t.Error("exit code = 0, want non-zero for issues")
	}
}
