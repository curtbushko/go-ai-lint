// Package testlint provides test data for the testlint analyzer.
package testlint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWithTestify is a test that properly uses testify.
// This should NOT trigger AIL130 because it imports testify.
func TestWithTestify(t *testing.T) {
	result := 2 + 2
	require.Equal(t, 4, result)
}

// TestAnotherWithTestify is another test with testify.
func TestAnotherWithTestify(t *testing.T) {
	require.True(t, true, "this should always pass")
}
