// Package testlint provides test data for the testlint analyzer.
package testlint

import "testing"

// TestWithoutTestify is a test that doesn't use testify.
// This should trigger AIL130.
func TestWithoutTestify(t *testing.T) { // want "AIL130: test file should use testify for assertions"
	if 1 != 1 {
		t.Fatal("unexpected")
	}
}

// TestAnotherWithoutTestify is another test without testify.
func TestAnotherWithoutTestify(t *testing.T) { // want "AIL130: test file should use testify for assertions"
	result := 2 + 2
	if result != 4 {
		t.Errorf("expected 4, got %d", result)
	}
}
