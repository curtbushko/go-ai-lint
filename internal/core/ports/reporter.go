package ports

import "github.com/curtbushko/go-ai-lint/internal/core/domain"

// Reporter defines the interface for outputting issues.
type Reporter interface {
	// Report outputs the given issues.
	Report(issues []domain.Issue) error
}

// Format specifies the output format for reporters.
type Format string

const (
	// FormatText is human-readable text output.
	FormatText Format = "text"
	// FormatJSON is machine-parseable JSON output.
	FormatJSON Format = "json"
	// FormatAI is AI-optimized output with guidance.
	FormatAI Format = "ai"
	// FormatSARIF is SARIF format for IDE integration.
	FormatSARIF Format = "sarif"
)
