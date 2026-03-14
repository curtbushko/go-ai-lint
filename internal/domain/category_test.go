package domain_test

import (
	"testing"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestCategoryConstants(t *testing.T) {
	// Test that all category constants exist
	categories := []domain.Category{
		domain.CategoryDefer,
		domain.CategoryContext,
		domain.CategoryGoroutine,
		domain.CategoryError,
		domain.CategoryNil,
		domain.CategoryType,
		domain.CategoryInterface,
		domain.CategoryNaming,
		domain.CategorySlice,
		domain.CategoryString,
		domain.CategoryConcurrency,
		domain.CategoryPanic,
		domain.CategoryInit,
		domain.CategoryOption,
		domain.CategoryCmd,
		domain.CategoryTest,
		domain.CategoryIO,
	}

	// Verify they are all different
	seen := make(map[domain.Category]bool)
	for _, cat := range categories {
		if seen[cat] {
			t.Errorf("Duplicate category: %s", cat)
		}
		seen[cat] = true
	}
}

func TestCategoryString(t *testing.T) {
	tests := []struct {
		category domain.Category
		want     string
	}{
		{domain.CategoryDefer, "defer"},
		{domain.CategoryContext, "context"},
		{domain.CategoryGoroutine, "goroutine"},
		{domain.CategoryError, "error"},
		{domain.CategoryNil, "nil"},
		{domain.CategoryType, "type"},
		{domain.CategoryInterface, "interface"},
		{domain.CategoryNaming, "naming"},
		{domain.CategorySlice, "slice"},
		{domain.CategoryString, "string"},
		{domain.CategoryConcurrency, "concurrency"},
		{domain.CategoryPanic, "panic"},
		{domain.CategoryInit, "init"},
		{domain.CategoryOption, "option"},
		{domain.CategoryCmd, "cmd"},
		{domain.CategoryTest, "test"},
		{domain.CategoryIO, "io"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.category.String()
			if got != tt.want {
				t.Errorf("Category.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
