package reporters_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/go-ai-lint/internal/adapters/reporters"
	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestJSONReporter(t *testing.T) {
	issues := []domain.Issue{
		{
			ID:       testIssueID,
			Name:     "defer-in-loop",
			Category: domain.CategoryDefer,
			Severity: domain.SeverityCritical,
			Position: domain.Position{
				Filename: "service.go",
				Line:     42,
				Column:   3,
			},
			Message: "defer inside loop delays resource cleanup",
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewJSONReporter(&buf)

	err := reporter.Report(issues)
	require.NoError(t, err)

	// Parse the output
	var result []reporters.JSONIssue
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to parse JSON output")

	require.Len(t, result, 1, "Expected 1 issue")

	issue := result[0]
	assert.Equal(t, testIssueID, issue.ID)
	assert.Equal(t, "service.go", issue.File)
	assert.Equal(t, 42, issue.Line)
}

func TestJSONReporterEmptyIssues(t *testing.T) {
	var buf bytes.Buffer
	reporter := reporters.NewJSONReporter(&buf)

	err := reporter.Report([]domain.Issue{})
	require.NoError(t, err)

	var result []reporters.JSONIssue
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to parse JSON output")

	assert.Empty(t, result, "Expected empty array")
}
