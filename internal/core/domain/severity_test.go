package domain_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
)

func TestSeverityConstants(t *testing.T) {
	// Test that severity constants exist and have correct ordering
	// Critical (0) > High (1) > Medium (2) > Low (3)
	tests := []struct {
		name     string
		severity domain.Severity
		wantVal  int
	}{
		{"Critical is 0", domain.SeverityCritical, 0},
		{"High is 1", domain.SeverityHigh, 1},
		{"Medium is 2", domain.SeverityMedium, 2},
		{"Low is 3", domain.SeverityLow, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.severity) != tt.wantVal {
				t.Errorf("got %d, want %d", tt.severity, tt.wantVal)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity domain.Severity
		want     string
	}{
		{domain.SeverityCritical, "critical"},
		{domain.SeverityHigh, "high"},
		{domain.SeverityMedium, "medium"},
		{domain.SeverityLow, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSeverityOrdering(t *testing.T) {
	// Critical should be "higher" severity than others (lower numeric value)
	if domain.SeverityCritical >= domain.SeverityHigh {
		t.Error("Critical should have lower value than High")
	}
	if domain.SeverityHigh >= domain.SeverityMedium {
		t.Error("High should have lower value than Medium")
	}
	if domain.SeverityMedium >= domain.SeverityLow {
		t.Error("Medium should have lower value than Low")
	}
}
