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

const testIssueID = "AIL001"

func TestAIReporter(t *testing.T) {
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
			Why:     "Deferred calls accumulate until function returns.",
			Fix:     "Extract loop body to separate function.",
			Example: domain.FixExample{
				Bad:         "for { defer f.Close() }",
				Good:        "for { processFile(f) }",
				Explanation: "Helper function scopes the defer.",
			},
			CommonMistakes: []string{
				"WRONG: Removing defer entirely",
				"WRONG: Moving defer outside loop",
			},
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewAIReporter(&buf)

	err := reporter.Report(issues)
	require.NoError(t, err)

	// Parse the output
	var result []reporters.AIIssue
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to parse JSON output: %s", buf.String())

	require.Len(t, result, 1, "Expected 1 issue")

	issue := result[0]

	// Check basic fields
	assert.Equal(t, testIssueID, issue.ID)

	// Check AI guidance fields
	assert.NotEmpty(t, issue.Why, "Why should not be empty")
	assert.NotEmpty(t, issue.Fix, "Fix should not be empty")
	assert.NotEmpty(t, issue.Example.Bad, "Example.Bad should not be empty")
	assert.NotEmpty(t, issue.Example.Good, "Example.Good should not be empty")
	assert.NotEmpty(t, issue.CommonMistakes, "CommonMistakes should not be empty")
}

func TestAIReporterContainsGuidance(t *testing.T) {
	issues := []domain.Issue{
		{
			ID:       testIssueID,
			Name:     "defer-in-loop",
			Category: domain.CategoryDefer,
			Severity: domain.SeverityCritical,
			Position: domain.Position{Filename: "test.go", Line: 1, Column: 1},
			Message:  "test message",
			Why:      "This is why it's a problem",
			Fix:      "This is how to fix it",
			CommonMistakes: []string{
				"WRONG: Do not do this",
			},
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewAIReporter(&buf)
	_ = reporter.Report(issues)

	output := buf.String()

	// Verify all guidance fields are present in output
	checks := []string{
		`"why"`,
		`"fix"`,
		`"common_mistakes"`,
		"This is why it's a problem",
		"This is how to fix it",
		"WRONG: Do not do this",
	}

	for _, check := range checks {
		assert.True(t, bytes.Contains(buf.Bytes(), []byte(check)), "Output missing %q\nGot: %s", check, output)
	}
}
