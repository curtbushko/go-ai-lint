// Package testlint provides test data for the testlint analyzer.
package testlint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupWithAssertNoError uses assert.NoError in setup position.
// This should trigger AIL132 because errors in setup should use require.
func TestSetupWithAssertNoError(t *testing.T) {
	err := setupSomething()
	assert.NoError(t, err) // want "AIL132: use require.NoError instead of assert.NoError for setup errors"

	// Continue with test that depends on the setup succeeding
	doSomethingElse()
}

// TestSetupWithAssertNotNil uses assert.NotNil in setup position.
// This should trigger AIL132 because nil checks in setup should use require.
func TestSetupWithAssertNotNil(t *testing.T) {
	obj := createObject()
	assert.NotNil(t, obj) // want "AIL132: use require.NotNil instead of assert.NotNil for setup checks"

	// Use the object - if it's nil this will panic
	useObject(obj)
}

// TestSetupWithAssertNil uses assert.Nil in setup position to check err.
// This should trigger AIL132.
func TestSetupWithAssertNil(t *testing.T) {
	err := setupSomething()
	assert.Nil(t, err) // want "AIL132: use require.Nil instead of assert.Nil for setup errors"

	doSomethingElse()
}

// TestRequireNoError uses require.NoError which is correct for setup.
// This should NOT trigger AIL132.
func TestRequireNoError(t *testing.T) {
	err := setupSomething()
	require.NoError(t, err)

	doSomethingElse()
}

// TestRequireNotNil uses require.NotNil which is correct for setup.
// This should NOT trigger AIL132.
func TestRequireNotNil(t *testing.T) {
	obj := createObject()
	require.NotNil(t, obj)

	useObject(obj)
}

// TestAssertEqualValidation uses assert.Equal for validation, not setup.
// This should NOT trigger AIL132.
func TestAssertEqualValidation(t *testing.T) {
	result := calculate()
	assert.Equal(t, 42, result) // This is validation, not setup
}

// TestAssertTrueValidation uses assert.True for validation.
// This should NOT trigger AIL132.
func TestAssertTrueValidation(t *testing.T) {
	flag := checkCondition()
	assert.True(t, flag) // This is validation, not setup
}

// TestMultipleAssertions has multiple assertions including setup assertions.
// Only the setup assertions should trigger AIL132.
func TestMultipleAssertions(t *testing.T) {
	err := setupSomething()
	assert.NoError(t, err) // want "AIL132: use require.NoError instead of assert.NoError for setup errors"

	obj := createObject()
	assert.NotNil(t, obj) // want "AIL132: use require.NotNil instead of assert.NotNil for setup checks"

	// These are validation assertions, should NOT trigger AIL132
	result := calculate()
	assert.Equal(t, 42, result)
	assert.True(t, result > 0)
}

// TestSubtestWithSetupAssert has subtests with setup assertions.
func TestSubtestWithSetupAssert(t *testing.T) {
	t.Run("subtest", func(t *testing.T) {
		err := setupSomething()
		assert.NoError(t, err) // want "AIL132: use require.NoError instead of assert.NoError for setup errors"
		doSomethingElse()
	})
}

// Helper functions for test compilation
func setupSomething() error {
	return nil
}

func doSomethingElse() {}

func createObject() interface{} {
	return struct{}{}
}

func useObject(obj interface{}) {}

func calculate() int {
	return 42
}

func checkCondition() bool {
	return true
}
