package ports_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "test", mock.Name())
}

func TestAnalyzerMethods(t *testing.T) {
	mock := &mockAnalyzer{name: "testlint"}

	assert.Equal(t, "testlint", mock.Name())

	analyzer := mock.Analyzer()
	require.NotNil(t, analyzer, "Analyzer() returned nil")
	assert.Equal(t, "testlint", analyzer.Name)
}
