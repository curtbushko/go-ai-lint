package stringlint

import "strings"

// ===== AIL070: string-byte-iteration =====

// BadByteIteration iterates over string bytes instead of runes.
func BadByteIteration(s string) {
	for i := 0; i < len(s); i++ { // want "AIL070: byte iteration on string"
		_ = s[i]
	}
}

// BadByteIterationWithBody iterates over string bytes with more complex body.
func BadByteIterationWithBody(s string) {
	for i := 0; i < len(s); i++ { // want "AIL070: byte iteration on string"
		b := s[i]
		if b == 'a' {
			continue
		}
	}
}

// GoodRuneIteration uses range to iterate over runes.
func GoodRuneIteration(s string) {
	for _, r := range s {
		_ = r
	}
}

// GoodByteSliceIteration iterates over []byte which is OK.
func GoodByteSliceIteration(b []byte) {
	for i := 0; i < len(b); i++ {
		_ = b[i]
	}
}

// GoodIndexIteration uses index for slicing, not byte access.
func GoodIndexIteration(s string) {
	for i := 0; i < len(s); i++ {
		_ = i // Using index only, not s[i]
	}
}

// ===== AIL071: string-concat-in-loop =====

// BadStringConcatInLoop uses += for string concatenation in loop.
func BadStringConcatInLoop(items []string) string {
	var result string
	for _, s := range items { // want "AIL071: string concatenation in loop"
		result += s
	}
	return result
}

// BadStringConcatInLoopWithSeparator uses += with separator.
func BadStringConcatInLoopWithSeparator(items []string) string {
	var result string
	for i, s := range items { // want "AIL071: string concatenation in loop"
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

// BadStringConcatInForLoop uses += in a traditional for loop.
func BadStringConcatInForLoop(n int) string {
	var result string
	for i := 0; i < n; i++ { // want "AIL071: string concatenation in loop"
		result += "x"
	}
	return result
}

// GoodStringBuilder uses strings.Builder.
func GoodStringBuilder(items []string) string {
	var b strings.Builder
	for _, s := range items {
		b.WriteString(s)
	}
	return b.String()
}

// GoodStringJoin uses strings.Join.
func GoodStringJoin(items []string) string {
	return strings.Join(items, "")
}

// GoodIntConcatInLoop concatenates ints, not strings.
func GoodIntConcatInLoop(items []int) int {
	var result int
	for _, n := range items {
		result += n
	}
	return result
}

// GoodSingleConcat is not in a loop.
func GoodSingleConcat(a, b string) string {
	result := a
	result += b
	return result
}
