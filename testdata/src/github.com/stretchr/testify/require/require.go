// Package require is a stub for the testify require package.
// This is used for testing the testlint analyzer's good case (tests with testify).
package require

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// Equal is a stub for require.Equal.
func Equal(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {}

// True is a stub for require.True.
func True(t TestingT, value bool, msgAndArgs ...interface{}) {}

// NoError is a stub for require.NoError.
func NoError(t TestingT, err error, msgAndArgs ...interface{}) {}
