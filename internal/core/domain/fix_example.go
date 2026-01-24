package domain

// FixExample provides concrete before/after code examples for an issue.
type FixExample struct {
	// Bad is the problematic code pattern.
	Bad string
	// Good is the correct code pattern.
	Good string
	// Explanation describes why the good version works.
	Explanation string
}
