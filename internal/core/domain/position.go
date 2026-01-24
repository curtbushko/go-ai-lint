package domain

import "fmt"

// Position represents a location in source code.
type Position struct {
	Filename  string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// String returns a string representation in the format "file:line:column".
func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}
