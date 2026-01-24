package errorlint

import (
	"errors"
	"fmt"
	"log"
)

// BadErrorHandledTwice demonstrates logging and returning an error.
func BadErrorHandledTwice() error {
	err := doSomething()
	if err != nil { // want "AIL030: error handled twice"
		log.Printf("error: %v", err)
		return err
	}
	return nil
}

// BadErrorHandledTwiceWithLogError uses log.Println.
func BadErrorHandledTwiceWithLogError() error {
	err := doSomething()
	if err != nil { // want "AIL030: error handled twice"
		log.Println("error occurred:", err)
		return err
	}
	return nil
}

// BadErrorHandledTwiceWrapped logs and returns wrapped error.
func BadErrorHandledTwiceWrapped() error {
	err := doSomething()
	if err != nil { // want "AIL030: error handled twice"
		log.Printf("failed: %v", err)
		return fmt.Errorf("operation failed: %w", err)
	}
	return nil
}

// BadErrorFmtNotWrapped uses %v instead of %w.
func BadErrorFmtNotWrapped() error {
	err := doSomething()
	if err != nil {
		return fmt.Errorf("operation failed: %v", err) // want "AIL033: error wrapped with %v instead of %w"
	}
	return nil
}

// BadErrorFmtNotWrappedMultiple has multiple format args.
func BadErrorFmtNotWrappedMultiple() error {
	err := doSomething()
	if err != nil {
		return fmt.Errorf("operation %s failed: %v", "test", err) // want "AIL033: error wrapped with %v instead of %w"
	}
	return nil
}

// GoodErrorLogOnly logs but doesn't return the error.
func GoodErrorLogOnly() {
	err := doSomething()
	if err != nil {
		log.Printf("error: %v", err)
		// No return of error - logging is the only handling
	}
}

// GoodErrorReturnOnly returns error without logging.
func GoodErrorReturnOnly() error {
	err := doSomething()
	if err != nil {
		return err
	}
	return nil
}

// GoodErrorWrappedProperly uses %w for error wrapping.
func GoodErrorWrappedProperly() error {
	err := doSomething()
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}
	return nil
}

// GoodErrorNewError creates a new error (not wrapping).
func GoodErrorNewError() error {
	err := doSomething()
	if err != nil {
		return errors.New("operation failed")
	}
	return nil
}

// GoodErrorFmtNoError uses fmt.Errorf without an error variable.
func GoodErrorFmtNoError() error {
	if somethingWrong() {
		return fmt.Errorf("something went wrong with value: %d", 42)
	}
	return nil
}

// GoodErrorLogThenDifferentReturn logs one error, returns different.
func GoodErrorLogThenDifferentReturn() error {
	err := doSomething()
	if err != nil {
		log.Printf("warning: %v", err)
		// Return a different, sentinel error
		return ErrOperationFailed
	}
	return nil
}

// GoodErrorLogInDeferReturn logs in defer, returns in if.
func GoodErrorLogInDeferReturn() (err error) {
	defer func() {
		if err != nil {
			log.Printf("operation completed with error: %v", err)
		}
	}()

	err = doSomething()
	if err != nil {
		return err
	}
	return nil
}

// Helper declarations for test compilation.
func doSomething() error                  { return nil }
func somethingWrong() bool                { return false }

var ErrOperationFailed = errors.New("operation failed")
