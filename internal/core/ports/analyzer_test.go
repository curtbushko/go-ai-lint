package ports_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/ports"
	"golang.org/x/tools/go/analysis"
)

// mockAnalyzer implements the Analyzer interface for testing.
type mockAnalyzer struct {
	name string
}

func (m *mockAnalyzer) Name() string {
	return m.name
}

func (m *mockAnalyzer) Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: m.name,
		Doc:  "mock analyzer for testing",
		Run: func(_ *analysis.Pass) (any, error) {
			return nil, nil
		},
	}
}

func TestAnalyzerInterface(_ *testing.T) {
	// Test that the interface can be implemented
	var _ ports.Analyzer = &mockAnalyzer{name: "test"}
}

func TestAnalyzerMethods(t *testing.T) {
	mock := &mockAnalyzer{name: "testlint"}

	if mock.Name() != "testlint" {
		t.Errorf("Name() = %q, want testlint", mock.Name())
	}

	analyzer := mock.Analyzer()
	if analyzer == nil {
		t.Fatal("Analyzer() returned nil")
	}
	if analyzer.Name != "testlint" {
		t.Errorf("Analyzer().Name = %q, want testlint", analyzer.Name)
	}
}
