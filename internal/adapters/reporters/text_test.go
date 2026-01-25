package reporters_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/adapters/reporters"
	"github.com/curtbushko/go-ai-lint/internal/core/domain"
)

func TestTextReporter(t *testing.T) {
	issues := []domain.Issue{
		{
			ID:       "AIL001",
			Name:     "defer-in-loop",
			Category: domain.CategoryDefer,
			Severity: domain.SeverityCritical,
			Position: domain.Position{
				Filename: "service.go",
				Line:     42,
				Column:   3,
			},
			Message: "defer inside loop delays resource cleanup",
			Why:     "Deferred calls accumulate until function returns.",
			Fix:     "Extract loop body to separate function.",
		},
	}

	t.Run("basic output", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := reporters.NewTextReporter(&buf, false)

		err := reporter.Report(issues)
		if err != nil {
			t.Fatalf("Report() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "service.go:42:3") {
			t.Errorf("output missing position, got: %s", output)
		}
		if !strings.Contains(output, "AIL001") {
			t.Errorf("output missing ID, got: %s", output)
		}
		if !strings.Contains(output, "defer-in-loop") {
			t.Errorf("output missing name, got: %s", output)
		}
	})

	t.Run("verbose output", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := reporters.NewTextReporter(&buf, true)

		err := reporter.Report(issues)
		if err != nil {
			t.Fatalf("Report() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Why:") {
			t.Errorf("verbose output missing Why, got: %s", output)
		}
		if !strings.Contains(output, "Fix:") {
			t.Errorf("verbose output missing Fix, got: %s", output)
		}
	})
}

func TestTextReporterEmptyIssues(t *testing.T) {
	var buf bytes.Buffer
	reporter := reporters.NewTextReporter(&buf, false)

	err := reporter.Report([]domain.Issue{})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}
