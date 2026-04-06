package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	assert.Equal(t, "AIL001", template.ID, "ID")
	assert.Equal(t, domain.SeverityCritical, template.Severity, "Severity")
	assert.Equal(t, "defer-in-loop", template.Name, "Name")
	assert.Equal(t, domain.CategoryDefer, template.Category, "Category")
	assert.Equal(t, "defer inside loop delays resource cleanup", template.Message, "Message")
	assert.Equal(t, "Deferred calls accumulate", template.Why, "Why")
	assert.Equal(t, "Extract to helper function", template.Fix, "Fix")
	assert.Len(t, template.CommonMistakes, 2, "CommonMistakes length")
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

	assert.Equal(t, template.ID, issue.ID, "Issue.ID")
	assert.Equal(t, template.Name, issue.Name, "Issue.Name")
	assert.Equal(t, template.Severity, issue.Severity, "Issue.Severity")
	assert.Equal(t, template.Category, issue.Category, "Issue.Category")
	assert.Equal(t, testFilename, issue.Position.Filename, "Issue.Position.Filename")
	assert.Equal(t, template.Why, issue.Why, "Issue.Why")
	assert.Equal(t, 1.0, issue.Confidence, "Issue.Confidence")
}
