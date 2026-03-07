package reporters_test

import (
	"bytes"
	"encoding/json"
	"testing"

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
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	// Parse the output as SARIF
	var result reporters.SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse SARIF output: %v\nOutput was: %s", err, buf.String())
	}

	// Validate SARIF schema version
	if result.Version != "2.1.0" {
		t.Errorf("Version = %q, want 2.1.0", result.Version)
	}

	if result.Schema != "https://json.schemastore.org/sarif-2.1.0.json" {
		t.Errorf("Schema = %q, want https://json.schemastore.org/sarif-2.1.0.json", result.Schema)
	}

	// Validate runs
	if len(result.Runs) != 1 {
		t.Fatalf("Expected 1 run, got %d", len(result.Runs))
	}

	run := result.Runs[0]

	// Validate tool info
	if run.Tool.Driver.Name != "go-ai-lint" {
		t.Errorf("Tool name = %q, want go-ai-lint", run.Tool.Driver.Name)
	}

	// Validate rules
	if len(run.Tool.Driver.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(run.Tool.Driver.Rules))
	}

	rule := run.Tool.Driver.Rules[0]
	if rule.ID != "AIL001" {
		t.Errorf("Rule ID = %q, want AIL001", rule.ID)
	}
	if rule.Name != "defer-in-loop" {
		t.Errorf("Rule Name = %q, want defer-in-loop", rule.Name)
	}

	// Validate results
	if len(run.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(run.Results))
	}

	res := run.Results[0]
	if res.RuleID != "AIL001" {
		t.Errorf("Result RuleID = %q, want AIL001", res.RuleID)
	}
	if res.Level != "error" {
		t.Errorf("Result Level = %q, want error", res.Level)
	}
	if res.Message.Text != "defer inside loop delays resource cleanup" {
		t.Errorf("Result Message = %q, want 'defer inside loop delays resource cleanup'", res.Message.Text)
	}

	// Validate location
	if len(res.Locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(res.Locations))
	}

	loc := res.Locations[0]
	if loc.PhysicalLocation.ArtifactLocation.URI != "service.go" {
		t.Errorf("URI = %q, want service.go", loc.PhysicalLocation.ArtifactLocation.URI)
	}
	if loc.PhysicalLocation.Region.StartLine != 42 {
		t.Errorf("StartLine = %d, want 42", loc.PhysicalLocation.Region.StartLine)
	}
	if loc.PhysicalLocation.Region.StartColumn != 3 {
		t.Errorf("StartColumn = %d, want 3", loc.PhysicalLocation.Region.StartColumn)
	}
}

func TestSARIFReporterEmptyIssues(t *testing.T) {
	var buf bytes.Buffer
	reporter := reporters.NewSARIFReporter(&buf)

	err := reporter.Report([]domain.Issue{})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	var result reporters.SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse SARIF output: %v", err)
	}

	if len(result.Runs) != 1 {
		t.Fatalf("Expected 1 run, got %d", len(result.Runs))
	}

	if len(result.Runs[0].Results) != 0 {
		t.Errorf("Expected empty results, got %d items", len(result.Runs[0].Results))
	}
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
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	var result reporters.SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse SARIF output: %v", err)
	}

	run := result.Runs[0]

	// Should have 2 rules (unique by ID)
	if len(run.Tool.Driver.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}

	// Should have 2 results
	if len(run.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(run.Results))
	}
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
			_ = reporter.Report(issues)

			var result reporters.SARIFLog
			_ = json.Unmarshal(buf.Bytes(), &result)

			got := result.Runs[0].Results[0].Level
			if got != tt.want {
				t.Errorf("severity %v: level = %q, want %q", tt.severity, got, tt.want)
			}
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
	_ = reporter.Report(issues)

	var result reporters.SARIFLog
	_ = json.Unmarshal(buf.Bytes(), &result)

	rule := result.Runs[0].Tool.Driver.Rules[0]

	// Rule should have help with Why and Fix info
	if rule.Help.Text == "" {
		t.Error("Rule help text should not be empty")
	}

	// Help should contain Why and Fix information
	if rule.FullDescription.Text == "" {
		t.Error("Rule full description should not be empty")
	}
}
