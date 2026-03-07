package domain_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestIssueFields(t *testing.T) {
	issue := domain.Issue{
		ID:         "AIL001",
		Name:       "defer-in-loop",
		Category:   domain.CategoryDefer,
		Severity:   domain.SeverityCritical,
		Position:   domain.Position{Filename: "service.go", Line: 42, Column: 3},
		Confidence: 1.0,
		Message:    "defer inside loop delays resource cleanup",
		Why:        "Deferred calls accumulate until function returns",
		Fix:        "Extract loop body to helper function",
		Example: domain.FixExample{
			Bad:         "for { defer f.Close() }",
			Good:        "for { processFile(f) }",
			Explanation: "Helper function runs defer after each iteration",
		},
		CommonMistakes: []string{
			"Removing defer entirely",
			"Moving defer outside loop",
		},
	}

	if issue.ID != "AIL001" {
		t.Errorf("ID = %q, want AIL001", issue.ID)
	}
	if issue.Name != "defer-in-loop" {
		t.Errorf("Name = %q, want defer-in-loop", issue.Name)
	}
	if issue.Category != domain.CategoryDefer {
		t.Errorf("Category = %s, want defer", issue.Category)
	}
	if issue.Severity != domain.SeverityCritical {
		t.Errorf("Severity = %s, want critical", issue.Severity)
	}
	if issue.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0", issue.Confidence)
	}
	if len(issue.CommonMistakes) != 2 {
		t.Errorf("CommonMistakes len = %d, want 2", len(issue.CommonMistakes))
	}
	if issue.Message != "defer inside loop delays resource cleanup" {
		t.Errorf("Message = %q, want correct message", issue.Message)
	}
	if issue.Why != "Deferred calls accumulate until function returns" {
		t.Errorf("Why = %q, want correct why", issue.Why)
	}
	if issue.Fix != "Extract loop body to helper function" {
		t.Errorf("Fix = %q, want correct fix", issue.Fix)
	}
}

func TestIssueString(t *testing.T) {
	issue := domain.Issue{
		ID:       "AIL001",
		Name:     "defer-in-loop",
		Position: domain.Position{Filename: "service.go", Line: 42, Column: 3},
		Message:  "defer inside loop delays resource cleanup",
	}

	got := issue.String()

	// Should contain position, ID, name, and message
	if !strings.Contains(got, "service.go:42:3") {
		t.Errorf("String() missing position, got %q", got)
	}
	if !strings.Contains(got, "AIL001") {
		t.Errorf("String() missing ID, got %q", got)
	}
	if !strings.Contains(got, "defer-in-loop") {
		t.Errorf("String() missing name, got %q", got)
	}
	if !strings.Contains(got, "defer inside loop") {
		t.Errorf("String() missing message, got %q", got)
	}
}

func TestNewIssue(t *testing.T) {
	pos := domain.Position{Filename: "test.go", Line: 10, Column: 5}

	issue := domain.NewIssue(
		"AIL002",
		"defer-close-error",
		domain.CategoryDefer,
		domain.SeverityHigh,
		pos,
		"deferred Close() error ignored",
	)

	if issue.ID != "AIL002" {
		t.Errorf("ID = %q, want AIL002", issue.ID)
	}
	if issue.Position.Filename != "test.go" {
		t.Errorf("Position.Filename = %q, want test.go", issue.Position.Filename)
	}
	if issue.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0 (default)", issue.Confidence)
	}
}
