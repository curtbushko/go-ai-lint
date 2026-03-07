package reporters

import (
	"encoding/json"
	"io"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

// JSONIssue is the JSON-serializable representation of an issue.
type JSONIssue struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
}

// JSONReporter outputs issues in JSON format.
type JSONReporter struct {
	w io.Writer
}

// NewJSONReporter creates a new JSON reporter.
func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{w: w}
}

// Report writes issues in JSON format.
func (r *JSONReporter) Report(issues []domain.Issue) error {
	jsonIssues := make([]JSONIssue, len(issues))
	for i, issue := range issues {
		jsonIssues[i] = JSONIssue{
			ID:       issue.ID,
			Name:     issue.Name,
			Category: string(issue.Category),
			Severity: issue.Severity.String(),
			File:     issue.Position.Filename,
			Line:     issue.Position.Line,
			Column:   issue.Position.Column,
			Message:  issue.Message,
		}
	}

	encoder := json.NewEncoder(r.w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonIssues)
}
