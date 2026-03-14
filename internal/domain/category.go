package domain

// Category represents the category of an issue.
type Category string

const (
	// CategoryDefer covers defer-related issues.
	CategoryDefer Category = "defer"
	// CategoryContext covers context misuse issues.
	CategoryContext Category = "context"
	// CategoryGoroutine covers goroutine lifecycle issues.
	CategoryGoroutine Category = "goroutine"
	// CategoryError covers error handling issues.
	CategoryError Category = "error"
	// CategoryNil covers nil-related issues.
	CategoryNil Category = "nil"
	// CategoryType covers type assertion issues.
	CategoryType Category = "type"
	// CategoryInterface covers interface design issues.
	CategoryInterface Category = "interface"
	// CategoryNaming covers naming convention issues.
	CategoryNaming Category = "naming"
	// CategorySlice covers slice and map issues.
	CategorySlice Category = "slice"
	// CategoryString covers string handling issues.
	CategoryString Category = "string"
	// CategoryConcurrency covers concurrency issues.
	CategoryConcurrency Category = "concurrency"
	// CategoryPanic covers panic/recover issues.
	CategoryPanic Category = "panic"
	// CategoryInit covers init function issues.
	CategoryInit Category = "init"
	// CategoryOption covers functional options pattern issues.
	CategoryOption Category = "option"
	// CategoryCmd covers CLI/command-related issues.
	CategoryCmd Category = "cmd"
)

// String returns the string representation of the category.
func (c Category) String() string {
	return string(c)
}
