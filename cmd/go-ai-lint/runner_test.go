package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, stdout.String(), "expected success message on stdout")
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

	// Then: Exit code is non-zero for issues
	assert.NotEqual(t, 0, exitCode, "exit code should be non-zero for issues")
}
