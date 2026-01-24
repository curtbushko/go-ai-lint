package domain_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
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
