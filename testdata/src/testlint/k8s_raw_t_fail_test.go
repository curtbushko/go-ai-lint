// Package testlint provides test data for the testlint analyzer.
package testlint

import (
	"testing"

	_ "k8s.io/kubernetes/test/e2e/framework"
)

// TestK8sWithRawTFail uses t.Errorf but is in a K8s e2e test file.
// This should NOT trigger AIL131 because K8s e2e tests are an exception.
func TestK8sWithRawTFail(t *testing.T) {
	result := 2 + 2
	if result != 5 {
		t.Errorf("expected 5, got %d", result) // No AIL131 expected due to k8s e2e exception
	}
}
