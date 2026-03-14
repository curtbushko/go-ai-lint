// Package testlint provides test data for the testlint analyzer.
package testlint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRawErrorf uses t.Errorf which should trigger AIL131.
func TestRawErrorf(t *testing.T) {
	result := 2 + 2
	if result != 5 {
		t.Errorf("expected 5, got %d", result) // want "AIL131: prefer testify assertions over raw t\\.Errorf/t\\.Fatalf/t\\.Error/t\\.Fatal"
	}
}

// TestRawFatalf uses t.Fatalf which should trigger AIL131.
func TestRawFatalf(t *testing.T) {
	result := 2 + 2
	if result != 5 {
		t.Fatalf("expected 5, got %d", result) // want "AIL131: prefer testify assertions over raw t\\.Errorf/t\\.Fatalf/t\\.Error/t\\.Fatal"
	}
}

// TestRawError uses t.Error which should trigger AIL131.
func TestRawError(t *testing.T) {
	result := 2 + 2
	if result != 5 {
		t.Error("unexpected result") // want "AIL131: prefer testify assertions over raw t\\.Errorf/t\\.Fatalf/t\\.Error/t\\.Fatal"
	}
}

// TestRawFatal uses t.Fatal which should trigger AIL131.
func TestRawFatal(t *testing.T) {
	result := 2 + 2
	if result != 5 {
		t.Fatal("unexpected result") // want "AIL131: prefer testify assertions over raw t\\.Errorf/t\\.Fatalf/t\\.Error/t\\.Fatal"
	}
}

// TestAssertEqual uses testify which should NOT trigger AIL131.
func TestAssertEqual(t *testing.T) {
	result := 2 + 2
	assert.Equal(t, 4, result)
}

// TestLog uses t.Log which should NOT trigger AIL131 (not an error method).
func TestLog(t *testing.T) {
	t.Log("this is a log message, not an error")
}
