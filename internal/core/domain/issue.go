package domain

import "fmt"

// Issue represents a detected code problem.
// Designed for both human and AI consumption.
type Issue struct {
	// ID is the unique identifier (e.g., "AIL001").
	ID string
	// Name is the short name (e.g., "defer-in-loop").
	Name string
	// Category is the issue category (Defer, Context, etc.).
	Category Category
	// Severity is the severity level.
	Severity Severity
	// Position is the file location.
	Position Position
	// Confidence is a 0.0-1.0 confidence level.
	Confidence float64

	// Message is a human-readable description of what was detected.
	Message string

	// Why explains the problem and its consequences (AI-consumable).
	Why string
	// Fix describes how to fix the issue (AI-consumable).
	Fix string
	// Example provides before/after code (AI-consumable).
	Example FixExample
	// CommonMistakes lists what NOT to do when fixing (AI-consumable).
	CommonMistakes []string
}

// String returns a formatted string representation of the issue.
func (i Issue) String() string {
	return fmt.Sprintf("%s: %s %s: %s", i.Position.String(), i.ID, i.Name, i.Message)
}

// NewIssue creates a new Issue with the given parameters and default confidence of 1.0.
func NewIssue(id, name string, category Category, severity Severity, pos Position, message string) Issue {
	return Issue{
		ID:         id,
		Name:       name,
		Category:   category,
		Severity:   severity,
		Position:   pos,
		Confidence: 1.0,
		Message:    message,
	}
}
