package deferlint

import (
	"bufio"
	"io"
	"os"
)

// BadDeferCloseErrorIgnored demonstrates ignoring Close() error.
func BadDeferCloseErrorIgnored(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // want "AIL002: deferred Close\\(\\) error is ignored"
	return nil
}

// BadDeferBodyCloseErrorIgnored demonstrates ignoring resp.Body.Close().
func BadDeferBodyCloseErrorIgnored(body io.ReadCloser) error {
	defer body.Close() // want "AIL002: deferred Close\\(\\) error is ignored"
	return nil
}

// BadDeferFlushErrorIgnored demonstrates ignoring Flush() error.
func BadDeferFlushErrorIgnored(w *bufio.Writer) error {
	defer w.Flush() // want "AIL003: deferred Flush\\(\\)/Sync\\(\\) error is ignored"
	return nil
}

// BadDeferSyncErrorIgnored demonstrates ignoring Sync() error.
func BadDeferSyncErrorIgnored(file *os.File) error {
	defer file.Sync() // want "AIL003: deferred Flush\\(\\)/Sync\\(\\) error is ignored"
	return nil
}

// GoodDeferCloseWithExplicitIgnore shows explicitly ignoring error.
func GoodDeferCloseWithExplicitIgnore(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }() // OK - explicitly ignored
	return nil
}

// GoodDeferCloseWithNamedReturn shows handling error via named return.
func GoodDeferCloseWithNamedReturn(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}() // OK - error is handled
	return nil
}

// GoodDeferCloseWithErrorCheck shows error check in defer.
func GoodDeferCloseWithErrorCheck(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log or handle error
			_ = err
		}
	}() // OK - error is checked
	return nil
}

// GoodReadOnlyClose shows Close on read-only file (less critical).
// Note: We still report this, but users can configure severity.
func GoodReadOnlyClose(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close() // want "AIL002: deferred Close\\(\\) error is ignored"
	return io.ReadAll(file)
}
