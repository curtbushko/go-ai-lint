package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestPositionFields(t *testing.T) {
	pos := domain.Position{
		Filename:  "test.go",
		Line:      42,
		Column:    10,
		EndLine:   42,
		EndColumn: 25,
	}

	assert.Equal(t, "test.go", pos.Filename, "Filename")
	assert.Equal(t, 42, pos.Line, "Line")
	assert.Equal(t, 10, pos.Column, "Column")
	assert.Equal(t, 42, pos.EndLine, "EndLine")
	assert.Equal(t, 25, pos.EndColumn, "EndColumn")
}

func TestPositionString(t *testing.T) {
	tests := []struct {
		name string
		pos  domain.Position
		want string
	}{
		{
			name: "standard position",
			pos: domain.Position{
				Filename: "service.go",
				Line:     42,
				Column:   3,
			},
			want: "service.go:42:3",
		},
		{
			name: "line 1 column 1",
			pos: domain.Position{
				Filename: "main.go",
				Line:     1,
				Column:   1,
			},
			want: "main.go:1:1",
		},
		{
			name: "path with directory",
			pos: domain.Position{
				Filename: "internal/core/domain/issue.go",
				Line:     100,
				Column:   15,
			},
			want: "internal/core/domain/issue.go:100:15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pos.String()
			assert.Equal(t, tt.want, got, "Position.String()")
		})
	}
}
