package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/curtbushko/go-ai-lint/internal/domain"
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
			assert.NotNil(t, tt.err, "error should not be nil")
			assert.NotEmpty(t, tt.err.Error(), "error message should not be empty")
		})
	}
}
