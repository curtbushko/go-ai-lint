package slicemaplint

// BadNilMapWrite demonstrates writing to nil map.
func BadNilMapWrite() {
	var m map[string]int
	m["key"] = 1 // want "AIL060: write to nil map will panic"
}

// BadNilMapWriteInFunc demonstrates nil map passed and written.
func BadNilMapWriteInFunc() {
	var m map[string]string
	populateMap(m) // The map is nil, but we can't detect this easily
}

func populateMap(m map[string]string) {
	// This is hard to detect statically without interprocedural analysis
	m["key"] = "value"
}

// BadNilMapWriteMultiple demonstrates multiple writes to nil map.
func BadNilMapWriteMultiple() {
	var counts map[string]int
	counts["a"] = 1 // want "AIL060: write to nil map will panic"
	counts["b"] = 2 // want "AIL060: write to nil map will panic"
}

// GoodMapWithMake demonstrates proper initialization.
func GoodMapWithMake() {
	m := make(map[string]int)
	m["key"] = 1 // OK - map is initialized
}

// GoodMapLiteral demonstrates map literal initialization.
func GoodMapLiteral() {
	m := map[string]int{}
	m["key"] = 1 // OK - map is initialized
}

// GoodMapLiteralWithValues demonstrates initialized map literal.
func GoodMapLiteralWithValues() {
	m := map[string]int{"existing": 0}
	m["key"] = 1 // OK - map is initialized
}

// GoodMapReassigned demonstrates reassigning nil map.
func GoodMapReassigned() {
	var m map[string]int
	m = make(map[string]int)
	m["key"] = 1 // OK - map was initialized before write
}

// GoodMapParameter demonstrates map received as parameter.
func GoodMapParameter(m map[string]int) {
	m["key"] = 1 // OK - caller's responsibility
}

// GoodMapFromFunction demonstrates map returned from function.
func GoodMapFromFunction() {
	m := createMap()
	m["key"] = 1 // OK - assuming function returns initialized map
}

func createMap() map[string]int {
	return make(map[string]int)
}

// GoodMapReadOnly demonstrates reading from nil map (doesn't panic).
func GoodMapReadOnly() {
	var m map[string]int
	_ = m["key"] // OK - reading from nil map returns zero value
}

// GoodMapNilCheck demonstrates checking for nil before write.
func GoodMapNilCheck() {
	var m map[string]int
	if m == nil {
		m = make(map[string]int)
	}
	m["key"] = 1 // OK - nil check before write
}
