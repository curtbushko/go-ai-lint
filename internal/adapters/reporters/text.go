// Package reporters provides output adapters for different formats.
package reporters

import (
	"fmt"
	"io"
	"strings"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

// TextReporter outputs issues in human-readable text format.
type TextReporter struct {
	w       io.Writer
	verbose bool
}

// NewTextReporter creates a new text reporter.
func NewTextReporter(w io.Writer, verbose bool) *TextReporter {
	return &TextReporter{w: w, verbose: verbose}
}

// Report writes issues in text format.
func (r *TextReporter) Report(issues []domain.Issue) error {
	for _, issue := range issues {
		// Basic format: file:line:col: ID message
		line := fmt.Sprintf("%s: %s %s: %s\n",
			issue.Position.String(),
			issue.ID,
			issue.Name,
			issue.Message,
		)
		if _, err := r.w.Write([]byte(line)); err != nil {
			return err
		}

		if r.verbose {
			r.writeVerbose(issue)
		}
	}
	return nil
}

func (r *TextReporter) writeVerbose(issue domain.Issue) {
	// Add why and fix information for verbose mode
	if issue.Why != "" {
		lines := strings.Split(issue.Why, "\n")
		for _, line := range lines {
			_, _ = fmt.Fprintf(r.w, "  Why: %s\n", line)
		}
	}
	if issue.Fix != "" {
		lines := strings.Split(issue.Fix, "\n")
		for _, line := range lines {
			_, _ = fmt.Fprintf(r.w, "  Fix: %s\n", line)
		}
	}
	_, _ = fmt.Fprintln(r.w)
}
