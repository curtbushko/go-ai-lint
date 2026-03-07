package reporters_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/adapters/reporters"
	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestJSONReporter(t *testing.T) {
	issues := []domain.Issue{
		{
			ID:       testIssueID,
			Name:     "defer-in-loop",
			Category: domain.CategoryDefer,
			Severity: domain.SeverityCritical,
			Position: domain.Position{
				Filename: "service.go",
				Line:     42,
				Column:   3,
			},
			Message: "defer inside loop delays resource cleanup",
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewJSONReporter(&buf)

	err := reporter.Report(issues)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	// Parse the output
	var result []reporters.JSONIssue
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result))
	}

	issue := result[0]
	if issue.ID != testIssueID {
		t.Errorf("ID = %q, want %s", issue.ID, testIssueID)
	}
	if issue.File != "service.go" {
		t.Errorf("File = %q, want service.go", issue.File)
	}
	if issue.Line != 42 {
		t.Errorf("Line = %d, want 42", issue.Line)
	}
}

func TestJSONReporterEmptyIssues(t *testing.T) {
	var buf bytes.Buffer
	reporter := reporters.NewJSONReporter(&buf)

	err := reporter.Report([]domain.Issue{})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	var result []reporters.JSONIssue
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty array, got %d items", len(result))
	}
}
