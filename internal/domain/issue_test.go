package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	assert.Equal(t, "AIL001", issue.ID, "ID")
	assert.Equal(t, "defer-in-loop", issue.Name, "Name")
	assert.Equal(t, domain.CategoryDefer, issue.Category, "Category")
	assert.Equal(t, domain.SeverityCritical, issue.Severity, "Severity")
	assert.Equal(t, 1.0, issue.Confidence, "Confidence")
	assert.Len(t, issue.CommonMistakes, 2, "CommonMistakes length")
	assert.Equal(t, "defer inside loop delays resource cleanup", issue.Message, "Message")
	assert.Equal(t, "Deferred calls accumulate until function returns", issue.Why, "Why")
	assert.Equal(t, "Extract loop body to helper function", issue.Fix, "Fix")
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
	assert.Contains(t, got, "service.go:42:3", "String() should contain position")
	assert.Contains(t, got, "AIL001", "String() should contain ID")
	assert.Contains(t, got, "defer-in-loop", "String() should contain name")
	assert.Contains(t, got, "defer inside loop", "String() should contain message")
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

	assert.Equal(t, "AIL002", issue.ID, "ID")
	assert.Equal(t, "test.go", issue.Position.Filename, "Position.Filename")
	assert.Equal(t, 1.0, issue.Confidence, "Confidence (default)")
}
