package ports_test

import (
	"testing"

	"golang.org/x/tools/go/analysis"

	"github.com/curtbushko/go-ai-lint/internal/ports"
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

func TestAnalyzerInterface(t *testing.T) {
	// Test that the interface can be implemented
	mock := &mockAnalyzer{name: "test"}
	var _ ports.Analyzer = mock
	if mock.Name() != "test" {
		t.Errorf("Name() = %q, want test", mock.Name())
	}
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
