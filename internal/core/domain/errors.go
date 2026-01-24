package domain

import "errors"

// Domain errors.
var (
	// ErrInvalidSeverity indicates an invalid severity value.
	ErrInvalidSeverity = errors.New("invalid severity value")
	// ErrInvalidCategory indicates an invalid category value.
	ErrInvalidCategory = errors.New("invalid category value")
)
