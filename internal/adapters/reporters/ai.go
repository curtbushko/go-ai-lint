package reporters

import (
	"encoding/json"
	"io"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

// AIIssue is the AI-optimized representation with full guidance.
type AIIssue struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`

	// Human-readable
	Message string `json:"message"`

	// AI-consumable guidance (prevents fix loops)
	Why            string    `json:"why"`
	Fix            string    `json:"fix"`
	Example        AIExample `json:"example,omitempty"`
	CommonMistakes []string  `json:"common_mistakes,omitempty"`
}

// AIExample provides before/after code examples.
type AIExample struct {
	Bad         string `json:"bad,omitempty"`
	Good        string `json:"good,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

// AIReporter outputs issues in AI-optimized format with full guidance.
// This format is designed to prevent AI assistants from getting stuck
// in fix loops by providing:
// - Why: The consequence of the problem
// - Fix: How to fix it (strategy)
// - Example: Before/after code
// - CommonMistakes: What NOT to do when fixing
type AIReporter struct {
	w io.Writer
}

// NewAIReporter creates a new AI reporter.
func NewAIReporter(w io.Writer) *AIReporter {
	return &AIReporter{w: w}
}

// Report writes issues in AI-optimized JSON format.
func (r *AIReporter) Report(issues []domain.Issue) error {
	aiIssues := make([]AIIssue, len(issues))
	for i, issue := range issues {
		aiIssues[i] = AIIssue{
			ID:       issue.ID,
			Name:     issue.Name,
			Category: string(issue.Category),
			Severity: issue.Severity.String(),
			File:     issue.Position.Filename,
			Line:     issue.Position.Line,
			Column:   issue.Position.Column,
			Message:  issue.Message,
			Why:      issue.Why,
			Fix:      issue.Fix,
			Example: AIExample{
				Bad:         issue.Example.Bad,
				Good:        issue.Example.Good,
				Explanation: issue.Example.Explanation,
			},
			CommonMistakes: issue.CommonMistakes,
		}
	}

	encoder := json.NewEncoder(r.w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aiIssues)
}
