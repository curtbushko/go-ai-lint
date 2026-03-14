// Package assert is a stub for the testify assert package.
// This is used for testing the testlint analyzer's good case (tests with testify).
package assert

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Errorf(format string, args ...interface{})
}

// Equal is a stub for assert.Equal.
func Equal(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return true
}

// True is a stub for assert.True.
func True(t TestingT, value bool, msgAndArgs ...interface{}) bool {
	return true
}

// NoError is a stub for assert.NoError.
func NoError(t TestingT, err error, msgAndArgs ...interface{}) bool {
	return true
}

// NotNil is a stub for assert.NotNil.
func NotNil(t TestingT, object interface{}, msgAndArgs ...interface{}) bool {
	return true
}

// Nil is a stub for assert.Nil.
func Nil(t TestingT, object interface{}, msgAndArgs ...interface{}) bool {
	return true
}
