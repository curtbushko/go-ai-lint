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

func TestSARIFReporter(t *testing.T) {
	issues := []domain.Issue{
		{
			ID:       "AIL001",
			Name:     "defer-in-loop",
			Category: domain.CategoryDefer,
			Severity: domain.SeverityCritical,
			Position: domain.Position{
				Filename:  "service.go",
				Line:      42,
				Column:    3,
				EndLine:   42,
				EndColumn: 10,
			},
			Message: "defer inside loop delays resource cleanup",
			Why:     "Defers accumulate and execute only when the function returns",
			Fix:     "Extract loop body to separate function",
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewSARIFReporter(&buf)

	err := reporter.Report(issues)
	require.NoError(t, err)

	// Parse the output as SARIF
	var result reporters.SARIFLog
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to parse SARIF output: %s", buf.String())

	// Validate SARIF schema version
	assert.Equal(t, "2.1.0", result.Version)
	assert.Equal(t, "https://json.schemastore.org/sarif-2.1.0.json", result.Schema)

	// Validate runs
	require.Len(t, result.Runs, 1, "Expected 1 run")

	run := result.Runs[0]

	// Validate tool info
	assert.Equal(t, "go-ai-lint", run.Tool.Driver.Name)

	// Validate rules
	require.Len(t, run.Tool.Driver.Rules, 1, "Expected 1 rule")

	rule := run.Tool.Driver.Rules[0]
	assert.Equal(t, "AIL001", rule.ID)
	assert.Equal(t, "defer-in-loop", rule.Name)

	// Validate results
	require.Len(t, run.Results, 1, "Expected 1 result")

	res := run.Results[0]
	assert.Equal(t, "AIL001", res.RuleID)
	assert.Equal(t, "error", res.Level)
	assert.Equal(t, "defer inside loop delays resource cleanup", res.Message.Text)

	// Validate location
	require.Len(t, res.Locations, 1, "Expected 1 location")

	loc := res.Locations[0]
	assert.Equal(t, "service.go", loc.PhysicalLocation.ArtifactLocation.URI)
	assert.Equal(t, 42, loc.PhysicalLocation.Region.StartLine)
	assert.Equal(t, 3, loc.PhysicalLocation.Region.StartColumn)
}

func TestSARIFReporterEmptyIssues(t *testing.T) {
	var buf bytes.Buffer
	reporter := reporters.NewSARIFReporter(&buf)

	err := reporter.Report([]domain.Issue{})
	require.NoError(t, err)

	var result reporters.SARIFLog
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to parse SARIF output")

	require.Len(t, result.Runs, 1, "Expected 1 run")
	assert.Empty(t, result.Runs[0].Results, "Expected empty results")
}

func TestSARIFReporterMultipleIssues(t *testing.T) {
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
			Message: "defer inside loop",
		},
		{
			ID:       "AIL002",
			Name:     "context-todo",
			Category: domain.CategoryContext,
			Severity: domain.SeverityMedium,
			Position: domain.Position{
				Filename: "handler.go",
				Line:     10,
				Column:   5,
			},
			Message: "context.TODO() should be replaced",
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewSARIFReporter(&buf)

	err := reporter.Report(issues)
	require.NoError(t, err)

	var result reporters.SARIFLog
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Failed to parse SARIF output")

	run := result.Runs[0]

	// Should have 2 rules (unique by ID)
	assert.Len(t, run.Tool.Driver.Rules, 2, "Expected 2 rules")

	// Should have 2 results
	assert.Len(t, run.Results, 2, "Expected 2 results")
}

func TestSARIFReporterSeverityMapping(t *testing.T) {
	tests := []struct {
		name     string
		severity domain.Severity
		want     string
	}{
		{"critical maps to error", domain.SeverityCritical, "error"},
		{"high maps to error", domain.SeverityHigh, "error"},
		{"medium maps to warning", domain.SeverityMedium, "warning"},
		{"low maps to note", domain.SeverityLow, "note"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := []domain.Issue{
				{
					ID:       "AIL001",
					Name:     "test-issue",
					Category: domain.CategoryDefer,
					Severity: tt.severity,
					Position: domain.Position{
						Filename: "test.go",
						Line:     1,
						Column:   1,
					},
					Message: "test message",
				},
			}

			var buf bytes.Buffer
			reporter := reporters.NewSARIFReporter(&buf)
			err := reporter.Report(issues)
			require.NoError(t, err)

			var result reporters.SARIFLog
			err = json.Unmarshal(buf.Bytes(), &result)
			require.NoError(t, err)

			got := result.Runs[0].Results[0].Level
			assert.Equal(t, tt.want, got, "severity %v", tt.severity)
		})
	}
}

func TestSARIFReporterRuleHelp(t *testing.T) {
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
			Why:     "Defers accumulate and execute only when the function returns",
			Fix:     "Extract loop body to separate function",
		},
	}

	var buf bytes.Buffer
	reporter := reporters.NewSARIFReporter(&buf)
	err := reporter.Report(issues)
	require.NoError(t, err)

	var result reporters.SARIFLog
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	rule := result.Runs[0].Tool.Driver.Rules[0]

	// Rule should have help with Why and Fix info
	assert.NotEmpty(t, rule.Help.Text, "Rule help text should not be empty")

	// Help should contain Why and Fix information
	assert.NotEmpty(t, rule.FullDescription.Text, "Rule full description should not be empty")
}
