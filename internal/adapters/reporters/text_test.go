package reporters_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/go-ai-lint/internal/adapters/reporters"
	"github.com/curtbushko/go-ai-lint/internal/domain"
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
		require.NoError(t, err)

		output := buf.String()
		assert.True(t, strings.Contains(output, "service.go:42:3"), "output missing position, got: %s", output)
		assert.True(t, strings.Contains(output, "AIL001"), "output missing ID, got: %s", output)
		assert.True(t, strings.Contains(output, "defer-in-loop"), "output missing name, got: %s", output)
	})

	t.Run("verbose output", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := reporters.NewTextReporter(&buf, true)

		err := reporter.Report(issues)
		require.NoError(t, err)

		output := buf.String()
		assert.True(t, strings.Contains(output, "Why:"), "verbose output missing Why, got: %s", output)
		assert.True(t, strings.Contains(output, "Fix:"), "verbose output missing Fix, got: %s", output)
	})
}

func TestTextReporterEmptyIssues(t *testing.T) {
	var buf bytes.Buffer
	reporter := reporters.NewTextReporter(&buf, false)

	err := reporter.Report([]domain.Issue{})
	require.NoError(t, err)

	assert.Equal(t, 0, buf.Len(), "expected empty output, got: %s", buf.String())
}
