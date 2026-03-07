// Package domain contains the core domain types for go-ai-lint.
// This package has no external dependencies - it is the innermost layer.
package domain

// Severity represents the severity level of an issue.
// Lower numeric values indicate higher severity.
type Severity int

const (
	// SeverityCritical indicates issues that will likely cause bugs or panics.
	SeverityCritical Severity = iota
	// SeverityHigh indicates issues that are likely problematic.
	SeverityHigh
	// SeverityMedium indicates code smells that should be fixed.
	SeverityMedium
	// SeverityLow indicates suggestions for improvement.
	SeverityLow
)

// String returns the string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "critical"
	case SeverityHigh:
		return "high"
	case SeverityMedium:
		return "medium"
	case SeverityLow:
		return "low"
	default:
		return "unknown"
	}
}
