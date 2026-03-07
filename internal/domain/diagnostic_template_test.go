package domain_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestDiagnosticTemplateFields(t *testing.T) {
	template := domain.DiagnosticTemplate{
		ID:       "AIL001",
		Name:     "defer-in-loop",
		Severity: domain.SeverityCritical,
		Category: domain.CategoryDefer,
		Message:  "defer inside loop delays resource cleanup",
		Why:      "Deferred calls accumulate",
		Fix:      "Extract to helper function",
		Example: domain.FixExample{
			Bad:  "bad code",
			Good: "good code",
		},
		CommonMistakes: []string{"mistake1", "mistake2"},
	}

	if template.ID != "AIL001" {
		t.Errorf("ID = %q, want AIL001", template.ID)
	}
	if template.Severity != domain.SeverityCritical {
		t.Errorf("Severity = %s, want critical", template.Severity)
	}
	if template.Name != "defer-in-loop" {
		t.Errorf("Name = %q, want defer-in-loop", template.Name)
	}
	if template.Category != domain.CategoryDefer {
		t.Errorf("Category = %s, want defer", template.Category)
	}
	if template.Message != "defer inside loop delays resource cleanup" {
		t.Errorf("Message = %q, want correct message", template.Message)
	}
	if template.Why != "Deferred calls accumulate" {
		t.Errorf("Why = %q, want correct why", template.Why)
	}
	if template.Fix != "Extract to helper function" {
		t.Errorf("Fix = %q, want correct fix", template.Fix)
	}
	if len(template.CommonMistakes) != 2 {
		t.Errorf("CommonMistakes len = %d, want 2", len(template.CommonMistakes))
	}
}

func TestDiagnosticTemplateCreateIssue(t *testing.T) {
	const testFilename = "test.go"

	template := domain.DiagnosticTemplate{
		ID:             "AIL001",
		Name:           "defer-in-loop",
		Severity:       domain.SeverityCritical,
		Category:       domain.CategoryDefer,
		Message:        "defer inside loop",
		Why:            "accumulates resources",
		Fix:            "use helper function",
		CommonMistakes: []string{"remove defer"},
	}

	pos := domain.Position{Filename: testFilename, Line: 10, Column: 5}
	issue := template.CreateIssue(pos)

	if issue.ID != template.ID {
		t.Errorf("Issue.ID = %q, want %q", issue.ID, template.ID)
	}
	if issue.Name != template.Name {
		t.Errorf("Issue.Name = %q, want %q", issue.Name, template.Name)
	}
	if issue.Severity != template.Severity {
		t.Errorf("Issue.Severity = %s, want %s", issue.Severity, template.Severity)
	}
	if issue.Category != template.Category {
		t.Errorf("Issue.Category = %s, want %s", issue.Category, template.Category)
	}
	if issue.Position.Filename != testFilename {
		t.Errorf("Issue.Position.Filename = %q, want %s", issue.Position.Filename, testFilename)
	}
	if issue.Why != template.Why {
		t.Errorf("Issue.Why = %q, want %q", issue.Why, template.Why)
	}
	if issue.Confidence != 1.0 {
		t.Errorf("Issue.Confidence = %f, want 1.0", issue.Confidence)
	}
}
