// Package testlint provides test data for the testlint analyzer.
package testlint

import (
	"testing"

	_ "sigs.k8s.io/e2e-framework/pkg/framework"
)

// TestSigsE2E is an e2e test that uses the sigs.k8s.io/e2e-framework.
// This should NOT trigger AIL130 because it imports the sigs e2e framework.
func TestSigsE2E(t *testing.T) {
	// Sigs e2e tests use the e2e-framework assertions
	// which are different from testify, but equally valid.
	// The analyzer should recognize this pattern and not flag it.
	t.Log("Running sigs e2e test")
}
