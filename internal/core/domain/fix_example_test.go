package domain_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/core/domain"
)

func TestFixExampleFields(t *testing.T) {
	example := domain.FixExample{
		Bad:         "defer file.Close()",
		Good:        "defer func() { _ = file.Close() }()",
		Explanation: "Capture the error from Close to avoid silent failures",
	}

	if example.Bad != "defer file.Close()" {
		t.Errorf("Bad = %q, unexpected", example.Bad)
	}
	if example.Good != "defer func() { _ = file.Close() }()" {
		t.Errorf("Good = %q, unexpected", example.Good)
	}
	if example.Explanation != "Capture the error from Close to avoid silent failures" {
		t.Errorf("Explanation = %q, unexpected", example.Explanation)
	}
}
