// Package main provides helper utilities for the go-ai-lint CLI.
package main

import "strings"

// parseCommaSeparated splits a comma-separated string into a slice.
// Empty strings between commas are ignored.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
