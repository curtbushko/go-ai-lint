package domain

// DiagnosticTemplate holds reusable diagnostic information for an issue type.
// Use CreateIssue to create an Issue instance at a specific position.
type DiagnosticTemplate struct {
	// ID is the unique identifier (e.g., "AIL001").
	ID string
	// Name is the short name (e.g., "defer-in-loop").
	Name string
	// Severity is the severity level.
	Severity Severity
	// Category is the issue category.
	Category Category
	// Message is the human-readable message template.
	Message string
	// Why explains the problem (AI-consumable).
	Why string
	// Fix describes how to fix (AI-consumable).
	Fix string
	// Example provides before/after code.
	Example FixExample
	// CommonMistakes lists what NOT to do when fixing.
	CommonMistakes []string
}

// CreateIssue creates an Issue from this template at the given position.
func (t DiagnosticTemplate) CreateIssue(pos Position) Issue {
	return Issue{
		ID:             t.ID,
		Name:           t.Name,
		Category:       t.Category,
		Severity:       t.Severity,
		Position:       pos,
		Confidence:     1.0,
		Message:        t.Message,
		Why:            t.Why,
		Fix:            t.Fix,
		Example:        t.Example,
		CommonMistakes: t.CommonMistakes,
	}
}
