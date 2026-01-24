package optionlint

// ===== AIL110: with-not-option =====

// Config is a sample configuration struct.
type Config struct {
	Timeout int
	Retries int
	Debug   bool
}

// BadWithTimeout doesn't follow functional options pattern.
func WithTimeout(timeout int) int { // want "AIL110: With.* function should return Option"
	return timeout
}

// BadWithRetries doesn't return an option function.
func WithRetries(retries int) string { // want "AIL110: With.* function should return Option"
	return "retries"
}

// BadWithDebug returns bool instead of option.
func WithDebug(debug bool) bool { // want "AIL110: With.* function should return Option"
	return debug
}

// Option is the functional option type.
type Option func(*Config)

// GoodWithTimeout returns an Option - proper pattern.
func GoodWithTimeout(timeout int) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// GoodWithRetries returns an Option - proper pattern.
func GoodWithRetries(retries int) Option {
	return func(c *Config) {
		c.Retries = retries
	}
}

// GoodWithDebug returns an Option - proper pattern.
func GoodWithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithContext is a common exception - often used differently.
func WithContext() {}

// WithValue is a common exception.
func WithValue() {}

// withdraw is not a With* function - different meaning.
func withdraw() {}

// Width is not a With* function.
func Width() int { return 0 }
