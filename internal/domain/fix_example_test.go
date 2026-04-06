package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestFixExampleFields(t *testing.T) {
	example := domain.FixExample{
		Bad:         "defer file.Close()",
		Good:        "defer func() { _ = file.Close() }()",
		Explanation: "Capture the error from Close to avoid silent failures",
	}

	assert.Equal(t, "defer file.Close()", example.Bad, "Bad")
	assert.Equal(t, "defer func() { _ = file.Close() }()", example.Good, "Good")
	assert.Equal(t, "Capture the error from Close to avoid silent failures", example.Explanation, "Explanation")
}
