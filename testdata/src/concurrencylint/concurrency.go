package concurrencylint

import "sync"

// ===== AIL080: waitgroup-done-not-deferred =====

// BadWaitGroupNotDeferred has wg.Done() not in defer.
func BadWaitGroupNotDeferred(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		doWork()
		wg.Done() // want "AIL080: wg.Done\\(\\) should be deferred"
	}()
}

// BadWaitGroupNotDeferredLocal creates WaitGroup locally.
func BadWaitGroupNotDeferredLocal() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		doWork()
		wg.Done() // want "AIL080: wg.Done\\(\\) should be deferred"
	}()
	wg.Wait()
}

// BadWaitGroupNotDeferredMultiple has multiple Done() calls not deferred.
func BadWaitGroupNotDeferredMultiple(wg *sync.WaitGroup) {
	wg.Add(2)
	go func() {
		if true {
			wg.Done() // want "AIL080: wg.Done\\(\\) should be deferred"
			return
		}
		wg.Done() // want "AIL080: wg.Done\\(\\) should be deferred"
	}()
}

// GoodWaitGroupDeferred uses defer wg.Done().
func GoodWaitGroupDeferred(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		doWork()
	}()
}

// GoodWaitGroupDeferredLocal creates WaitGroup locally with defer.
func GoodWaitGroupDeferredLocal() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		doWork()
	}()
	wg.Wait()
}

// GoodWaitGroupNotInGoroutine - Done() outside goroutine is not our concern.
func GoodWaitGroupNotInGoroutine(wg *sync.WaitGroup) {
	wg.Add(1)
	doWork()
	wg.Done() // OK - not in goroutine, pattern varies
}

// ===== AIL082: select-only-default =====

// BadSelectOnlyDefault has select with only default case.
func BadSelectOnlyDefault() {
	select { // want "AIL082: select with only default case is useless"
	default:
		doWork()
	}
}

// BadSelectOnlyDefaultEmpty has empty default.
func BadSelectOnlyDefaultEmpty() {
	select { // want "AIL082: select with only default case is useless"
	default:
	}
}

// GoodSelectWithCases has channel cases.
func GoodSelectWithCases(ch chan int, done chan struct{}) {
	select {
	case v := <-ch:
		_ = v
	case <-done:
		return
	default:
		doWork()
	}
}

// GoodSelectNoCases - empty select blocks forever (intentional pattern).
func GoodSelectNoCases() {
	// select {} // This is valid Go but blocks forever - we don't flag it
}

// GoodSelectSingleCase has one channel case.
func GoodSelectSingleCase(ch chan int) {
	select {
	case v := <-ch:
		_ = v
	}
}

// GoodSelectCaseAndDefault has case and default.
func GoodSelectCaseAndDefault(ch chan int) {
	select {
	case v := <-ch:
		_ = v
	default:
		doWork()
	}
}

// Helper declarations for test compilation.
func doWork() {}
