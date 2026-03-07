package reporters_test

import (
	"bytes"
	"encoding/json"
	"testing"

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
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	// Parse the output
	var result []reporters.AIIssue
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(result))
	}

	issue := result[0]

	// Check basic fields
	if issue.ID != testIssueID {
		t.Errorf("ID = %q, want %s", issue.ID, testIssueID)
	}

	// Check AI guidance fields
	if issue.Why == "" {
		t.Error("Why should not be empty")
	}
	if issue.Fix == "" {
		t.Error("Fix should not be empty")
	}
	if issue.Example.Bad == "" {
		t.Error("Example.Bad should not be empty")
	}
	if issue.Example.Good == "" {
		t.Error("Example.Good should not be empty")
	}
	if len(issue.CommonMistakes) == 0 {
		t.Error("CommonMistakes should not be empty")
	}
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
		if !bytes.Contains(buf.Bytes(), []byte(check)) {
			t.Errorf("Output missing %q\nGot: %s", check, output)
		}
	}
}
