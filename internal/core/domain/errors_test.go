package domain_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
)

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidSeverity exists", domain.ErrInvalidSeverity},
		{"ErrInvalidCategory exists", domain.ErrInvalidCategory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}
