package ports_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
)

// mockReporter implements the Reporter interface for testing.
type mockReporter struct {
	reported []domain.Issue
}

func (m *mockReporter) Report(issues []domain.Issue) error {
	m.reported = issues
	return nil
}

func TestReporterInterface(_ *testing.T) {
	// Test that the interface can be implemented
	var _ ports.Reporter = &mockReporter{}
}

func TestReporterReport(t *testing.T) {
	mock := &mockReporter{}

	issues := []domain.Issue{
		domain.NewIssue("AIL001", "test", domain.CategoryDefer, domain.SeverityHigh,
			domain.Position{Filename: "test.go", Line: 1, Column: 1}, "test message"),
	}

	err := mock.Report(issues)
	if err != nil {
		t.Errorf("Report() error = %v", err)
	}
	if len(mock.reported) != 1 {
		t.Errorf("reported %d issues, want 1", len(mock.reported))
	}
}

func TestFormatConstants(t *testing.T) {
	tests := []struct {
		format ports.Format
		want   string
	}{
		{ports.FormatText, "text"},
		{ports.FormatJSON, "json"},
		{ports.FormatAI, "ai"},
		{ports.FormatSARIF, "sarif"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.format) != tt.want {
				t.Errorf("Format = %q, want %q", tt.format, tt.want)
			}
		})
	}
}
