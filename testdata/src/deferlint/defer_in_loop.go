package deferlint

import "os"

// BadDeferInForLoop demonstrates defer inside a for loop.
func BadDeferInForLoop(files []string) error {
	for i := 0; i < len(files); i++ {
		file, err := os.Open(files[i])
		if err != nil {
			return err
		}
		defer file.Close() // want "AIL001: defer inside loop" "AIL002: deferred Close\\(\\) error is ignored"
	}
	return nil
}

// BadDeferInRangeLoop demonstrates defer inside a range loop.
func BadDeferInRangeLoop(files []string) error {
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return err
		}
		defer file.Close() // want "AIL001: defer inside loop" "AIL002: deferred Close\\(\\) error is ignored"
	}
	return nil
}

// BadDeferInNestedLoop demonstrates defer in nested loops.
func BadDeferInNestedLoop(dirs []string) error {
	for _, dir := range dirs {
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			file, err := os.Open(f.Name())
			if err != nil {
				return err
			}
			defer file.Close() // want "AIL001: defer inside loop" "AIL002: deferred Close\\(\\) error is ignored"
		}
	}
	return nil
}

// GoodDeferOutsideLoop demonstrates correct defer usage.
func GoodDeferOutsideLoop(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // want "AIL002: deferred Close\\(\\) error is ignored"
	return nil
}

// GoodDeferInHelperFunction demonstrates extracting to helper.
func GoodDeferInHelperFunction(files []string) error {
	for _, f := range files {
		if err := processFile(f); err != nil {
			return err
		}
	}
	return nil
}

func processFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close() // want "AIL002: deferred Close\\(\\) error is ignored"
	// process file...
	return nil
}

// GoodDeferInClosure demonstrates using immediately invoked closure.
func GoodDeferInClosure(files []string) error {
	for _, f := range files {
		if err := func() error {
			file, err := os.Open(f)
			if err != nil {
				return err
			}
			defer file.Close() // want "AIL002: deferred Close\\(\\) error is ignored"
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}
