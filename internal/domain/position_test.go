package domain_test

import (
	"testing"

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

	if pos.Filename != "test.go" {
		t.Errorf("Filename = %q, want %q", pos.Filename, "test.go")
	}
	if pos.Line != 42 {
		t.Errorf("Line = %d, want %d", pos.Line, 42)
	}
	if pos.Column != 10 {
		t.Errorf("Column = %d, want %d", pos.Column, 10)
	}
	if pos.EndLine != 42 {
		t.Errorf("EndLine = %d, want %d", pos.EndLine, 42)
	}
	if pos.EndColumn != 25 {
		t.Errorf("EndColumn = %d, want %d", pos.EndColumn, 25)
	}
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
			if got != tt.want {
				t.Errorf("Position.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
