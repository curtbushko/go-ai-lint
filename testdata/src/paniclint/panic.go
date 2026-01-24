package paniclint

import "errors"

// ===== AIL090: panic-in-library =====

// Config is a sample config struct.
type Config struct{}

// ParseConfig panics in library code - bad practice.
func ParseConfig(data []byte) Config {
	if len(data) == 0 {
		panic("empty config") // want "AIL090: panic in library code"
	}
	return Config{}
}

// ValidateInput panics instead of returning error.
func ValidateInput(s string) {
	if s == "" {
		panic("empty input") // want "AIL090: panic in library code"
	}
}

// ProcessData panics on invalid data.
func ProcessData(data []byte) []byte {
	if data == nil {
		panic(errors.New("nil data")) // want "AIL090: panic in library code"
	}
	return data
}

// MustParseConfig is OK - Must prefix indicates panic is expected.
func MustParseConfig(data []byte) Config {
	if len(data) == 0 {
		panic("empty config") // OK - Must prefix
	}
	return Config{}
}

// MustValidate is OK - Must prefix.
func MustValidate(s string) {
	if s == "" {
		panic("empty") // OK - Must prefix
	}
}

// ParseConfigSafe returns error properly - good practice.
func ParseConfigSafe(data []byte) (Config, error) {
	if len(data) == 0 {
		return Config{}, errors.New("empty config")
	}
	return Config{}, nil
}

// ===== AIL091: empty-recover =====

// BadEmptyRecover discards the recover result.
func BadEmptyRecover() {
	defer func() {
		recover() // want "AIL091: recover\\(\\) result discarded"
	}()
	doWork()
}

// BadEmptyRecoverBlank assigns to blank identifier.
func BadEmptyRecoverBlank() {
	defer func() {
		_ = recover() // want "AIL091: recover\\(\\) result discarded"
	}()
	doWork()
}

// GoodRecoverWithLog uses the recover result.
func GoodRecoverWithLog() {
	defer func() {
		if r := recover(); r != nil {
			logError(r)
		}
	}()
	doWork()
}

// GoodRecoverWithCheck checks the recover result.
func GoodRecoverWithCheck() {
	defer func() {
		if err := recover(); err != nil {
			handleError(err)
		}
	}()
	doWork()
}

// GoodRecoverWithRePanic re-panics after logging.
func GoodRecoverWithRePanic() {
	defer func() {
		if r := recover(); r != nil {
			logError(r)
			panic(r) // Re-panic after logging
		}
	}()
	doWork()
}

// Helper functions for compilation.
func doWork()         {}
func logError(any)    {}
func handleError(any) {}
