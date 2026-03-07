package goroutinelint

import (
	"context"
	"time"
)

// BadGoroutineNoCancel demonstrates goroutine without cancellation mechanism.
func BadGoroutineNoCancel() {
	go func() { // want "AIL020: goroutine without cancellation mechanism"
		for {
			doWork()
			time.Sleep(time.Second)
		}
	}()
}

// BadGoroutineInfiniteLoop demonstrates infinite loop without context.
func BadGoroutineInfiniteLoop() {
	go func() { // want "AIL021: infinite loop in goroutine without exit condition"
		for {
			doWork()
		}
	}()
}

// BadGoroutineClosureCapture demonstrates loop variable capture (pre-1.22).
func BadGoroutineClosureCapture() {
	items := []int{1, 2, 3}
	for _, item := range items {
		go func() { // want "AIL022: loop variable 'item' captured by goroutine"
			process(item)
		}()
	}
}

// BadGoroutineClosureCaptureIndex demonstrates index variable capture.
func BadGoroutineClosureCaptureIndex() {
	items := []int{1, 2, 3}
	for i := range items {
		go func() { // want "AIL022: loop variable 'i' captured by goroutine"
			processIndex(i)
		}()
	}
}

// GoodGoroutineWithContext demonstrates proper cancellation with context.
func GoodGoroutineWithContext(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				doWork()
			}
		}
	}()
}

// GoodGoroutineWithDoneChannel demonstrates cancellation via done channel.
func GoodGoroutineWithDoneChannel(done <-chan struct{}) {
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				doWork()
			}
		}
	}()
}

// GoodGoroutineWithBreak demonstrates loop with break condition.
func GoodGoroutineWithBreak() {
	go func() {
		for {
			if shouldStop() {
				break
			}
			doWork()
		}
	}()
}

// GoodGoroutineWithReturn demonstrates loop with return condition.
func GoodGoroutineWithReturn() {
	go func() {
		for {
			if shouldStop() {
				return
			}
			doWork()
		}
	}()
}

// GoodGoroutineFiniteLoop demonstrates finite for loop.
func GoodGoroutineFiniteLoop() {
	go func() {
		for i := 0; i < 10; i++ {
			doWork()
		}
	}()
}

// GoodGoroutineRangeLoop demonstrates range loop (finite).
func GoodGoroutineRangeLoop(items []int) {
	go func() {
		for _, item := range items {
			process(item)
		}
	}()
}

// GoodGoroutineClosureCaptureFixed demonstrates proper loop variable handling.
func GoodGoroutineClosureCaptureFixed() {
	items := []int{1, 2, 3}
	for _, item := range items {
		item := item // Shadow the loop variable
		go func() {
			process(item)
		}()
	}
}

// GoodGoroutineClosureCaptureParam demonstrates passing loop variable as parameter.
func GoodGoroutineClosureCaptureParam() {
	items := []int{1, 2, 3}
	for _, item := range items {
		go func(v int) {
			process(v)
		}(item)
	}
}

// GoodGoroutineNoLoop demonstrates goroutine without loop.
func GoodGoroutineNoLoop() {
	go func() {
		doWork()
	}()
}

// GoodGoroutineWithTimer demonstrates goroutine with timer-based exit.
func GoodGoroutineWithTimer() {
	go func() {
		timer := time.NewTimer(time.Minute)
		for {
			select {
			case <-timer.C:
				return
			default:
				doWork()
			}
		}
	}()
}

// Helper functions for test compilation.
func doWork()          {}
func process(int)      {}
func processIndex(int) {}
func shouldStop() bool { return false }
