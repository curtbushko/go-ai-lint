package interfacelint

// ===== AIL040: interface-too-large =====

// Repository is a god interface with too many methods.
type Repository interface { // want "AIL040: interface has 7 methods"
	Create(item Item) error
	Read(id string) (Item, error)
	Update(item Item) error
	Delete(id string) error
	List() ([]Item, error)
	Search(query string) ([]Item, error)
	Count() (int, error)
}

// BigInterface has 6 methods - also too large.
type BigInterface interface { // want "AIL040: interface has 6 methods"
	Method1()
	Method2()
	Method3()
	Method4()
	Method5()
	Method6()
}

// FiveMethodInterface has exactly 5 methods - borderline but OK.
type FiveMethodInterface interface {
	Method1()
	Method2()
	Method3()
	Method4()
	Method5()
}

// SmallInterface is a good small interface.
type SmallInterface interface {
	Method1()
	Method2()
}

// Reader is a good focused interface.
type Reader interface {
	Read(id string) (Item, error)
}

// Writer is a good focused interface.
type Writer interface {
	Write(item Item) error
}

// ===== AIL042: interface-missing-er-suffix =====

// Validate is a single-method interface without -er suffix.
type Validate interface { // want "AIL042: single-method interface should have"
	Validate() error
}

// Process is a single-method interface without -er suffix.
type Process interface { // want "AIL042: single-method interface should have"
	Process() error
}

// Handle is a single-method interface without -er suffix.
type Handle interface { // want "AIL042: single-method interface should have"
	Handle() error
}

// Validator is a good single-method interface with -er suffix.
type Validator interface {
	Validate() error
}

// Processor is a good single-method interface with -er suffix.
type Processor interface {
	Process() error
}

// Handler is a good single-method interface with -er suffix.
type Handler interface {
	Handle() error
}

// Closer is a good single-method interface with -er suffix.
type Closer interface {
	Close() error
}

// TwoMethodInterface has two methods - no -er suffix required.
type TwoMethodInterface interface {
	Method1()
	Method2()
}

// EmptyInterface is empty - no -er suffix required.
type EmptyInterface interface{}

// ===== nolint directive tests =====

// SuppressedLargeInterface is suppressed with nolint directive on same line.
type SuppressedLargeInterface interface { //nolint:interfacelint
	Method1()
	Method2()
	Method3()
	Method4()
	Method5()
	Method6()
}

// SuppressedValidate is suppressed with nolint on line above.
//
//nolint:interfacelint
type SuppressedValidate interface {
	Validate() error
}

// SuppressedAll is suppressed with nolint (all analyzers).
type SuppressedAll interface { //nolint
	Process() error
}

// Helper type for compilation.
type Item struct{}
