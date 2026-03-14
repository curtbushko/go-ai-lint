// Package testlint provides test data for the testlint analyzer.
package testlint

import (
	"testing"

	_ "k8s.io/kubernetes/test/e2e/framework"
)

// TestK8sE2E is an e2e test that uses Kubernetes framework conventions.
// This should NOT trigger AIL130 because it imports the k8s e2e framework.
func TestK8sE2E(t *testing.T) {
	// Kubernetes e2e tests typically use framework assertions
	// which are different from testify, but equally valid.
	// The analyzer should recognize this pattern and not flag it.
	t.Log("Running k8s e2e test")
}
