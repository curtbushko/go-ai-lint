package domain_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

func TestNolintEnabledRespectsDirectives(t *testing.T) {
	// Given: Config has nolint.enabled: true (default)
	domain.SetNolintEnabled(true)
	defer domain.SetNolintEnabled(true) // Reset after test

	code := `package test
type Validate interface { //nolint:interfacelint
	Validate() error
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err, "failed to parse code")

	// Find position on line 2 (the interface declaration)
	var targetPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == 2 {
			targetPos = n.Pos()
			return false
		}
		return true
	})

	// When: Check if should skip with nolint directive
	// Then: Issue is suppressed by nolint directive
	got := domain.ShouldSkip(fset, file.Comments, targetPos, "interfacelint")
	assert.True(t, got, "ShouldSkip() should be true when nolint is enabled")
}

func TestNolintDisabledIgnoresDirectives(t *testing.T) {
	// Given: Config has nolint.enabled: false
	domain.SetNolintEnabled(false)
	defer domain.SetNolintEnabled(true) // Reset after test

	code := `package test
type Validate interface { //nolint:interfacelint
	Validate() error
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err, "failed to parse code")

	// Find position on line 2 (the interface declaration)
	var targetPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == 2 {
			targetPos = n.Pos()
			return false
		}
		return true
	})

	// When: Check if should skip with nolint directive but nolint disabled
	// Then: Issue is still reported (directive ignored)
	got := domain.ShouldSkip(fset, file.Comments, targetPos, "interfacelint")
	assert.False(t, got, "ShouldSkip() should be false when nolint is disabled")
}

func TestNolintSettingGlobal(t *testing.T) {
	// Given: Config has nolint.enabled: false
	domain.SetNolintEnabled(false)
	defer domain.SetNolintEnabled(true) // Reset after test

	// Test with multiple analyzer types to ensure global setting applies to all
	testCases := []struct {
		name         string
		analyzerName string
		code         string
	}{
		{
			name:         "interfacelint",
			analyzerName: "interfacelint",
			code: `package test
type Validate interface { //nolint:interfacelint
	Validate() error
}`,
		},
		{
			name:         "errorlint",
			analyzerName: "errorlint",
			code: `package test
func foo() error { //nolint:errorlint
	return nil
}`,
		},
		{
			name:         "nolint all",
			analyzerName: "deferlint",
			code: `package test
func foo() { //nolint
	defer func() {}()
}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tc.code, parser.ParseComments)
			require.NoError(t, err, "failed to parse code")

			// Find position on line 2
			var targetPos token.Pos
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				pos := fset.Position(n.Pos())
				if pos.Line == 2 {
					targetPos = n.Pos()
					return false
				}
				return true
			})

			// When: Run analyzer on code with nolint directives
			// Then: All analyzers report issues (none respect nolint)
			got := domain.ShouldSkip(fset, file.Comments, targetPos, tc.analyzerName)
			assert.False(t, got, "ShouldSkip() should be false for %s when nolint is disabled globally", tc.analyzerName)
		})
	}
}

func TestParseDirective(t *testing.T) {
	tests := []struct {
		name     string
		comment  string
		wantAll  bool
		wantList []string
	}{
		{
			name:     "nolint all",
			comment:  "//nolint",
			wantAll:  true,
			wantList: nil,
		},
		{
			name:     "nolint with space",
			comment:  "// nolint",
			wantAll:  true,
			wantList: nil,
		},
		{
			name:     "nolint single analyzer",
			comment:  "//nolint:interfacelint",
			wantAll:  false,
			wantList: []string{"interfacelint"},
		},
		{
			name:     "nolint multiple analyzers",
			comment:  "//nolint:interfacelint,errorlint",
			wantAll:  false,
			wantList: []string{"interfacelint", "errorlint"},
		},
		{
			name:     "nolint with spaces",
			comment:  "//nolint: interfacelint, errorlint",
			wantAll:  false,
			wantList: []string{"interfacelint", "errorlint"},
		},
		{
			name:     "not a nolint comment",
			comment:  "// regular comment",
			wantAll:  false,
			wantList: nil,
		},
		{
			name:     "empty comment",
			comment:  "",
			wantAll:  false,
			wantList: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAll, gotList := domain.ParseDirective(tt.comment)
			assert.Equal(t, tt.wantAll, gotAll, "ParseDirective(%q) all", tt.comment)
			assert.Equal(t, tt.wantList, gotList, "ParseDirective(%q) list", tt.comment)
		})
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		analyzerName string
		lineToCheck  int // 1-indexed line number to check
		want         bool
	}{
		{
			name: "nolint all on same line",
			code: `package test
type Validate interface { //nolint
	Validate() error
}`,
			analyzerName: "interfacelint",
			lineToCheck:  2,
			want:         true,
		},
		{
			name: "nolint specific analyzer on same line",
			code: `package test
type Validate interface { //nolint:interfacelint
	Validate() error
}`,
			analyzerName: "interfacelint",
			lineToCheck:  2,
			want:         true,
		},
		{
			name: "nolint different analyzer on same line",
			code: `package test
type Validate interface { //nolint:errorlint
	Validate() error
}`,
			analyzerName: "interfacelint",
			lineToCheck:  2,
			want:         false,
		},
		{
			name: "nolint on line above",
			code: `package test
//nolint:interfacelint
type Validate interface {
	Validate() error
}`,
			analyzerName: "interfacelint",
			lineToCheck:  3,
			want:         true,
		},
		{
			name: "no nolint comment",
			code: `package test
type Validate interface {
	Validate() error
}`,
			analyzerName: "interfacelint",
			lineToCheck:  2,
			want:         false,
		},
		{
			name: "nolint multiple analyzers includes target",
			code: `package test
type Validate interface { //nolint:errorlint,interfacelint,paniclint
	Validate() error
}`,
			analyzerName: "interfacelint",
			lineToCheck:  2,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err, "failed to parse code")

			// Find position on the target line
			var targetPos token.Pos
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				pos := fset.Position(n.Pos())
				if pos.Line == tt.lineToCheck {
					targetPos = n.Pos()
					return false
				}
				return true
			})

			require.NotEqual(t, token.NoPos, targetPos, "could not find node on line %d", tt.lineToCheck)

			got := domain.ShouldSkip(fset, file.Comments, targetPos, tt.analyzerName)
			assert.Equal(t, tt.want, got, "ShouldSkip()")
		})
	}
}

func TestRequireSpecificDisabledAllowsBare(t *testing.T) {
	// Given: Config has nolint.require-specific: false (default)
	domain.SetNolintEnabled(true)
	domain.SetRequireSpecific(false)
	defer func() {
		domain.SetNolintEnabled(true)
		domain.SetRequireSpecific(false)
	}()

	code := `package test
func foo() { //nolint
	defer func() {}()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err, "failed to parse code")

	// Find position on line 2
	var targetPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == 2 {
			targetPos = n.Pos()
			return false
		}
		return true
	})

	// When: Analyze code with bare //nolint comment
	// Then: Issue is suppressed
	got := domain.ShouldSkip(fset, file.Comments, targetPos, "deferlint")
	assert.True(t, got, "ShouldSkip() should be true when require-specific is false")
}

func TestRequireSpecificRejectsBare(t *testing.T) {
	// Given: Config has nolint.require-specific: true
	domain.SetNolintEnabled(true)
	domain.SetRequireSpecific(true)
	defer func() {
		domain.SetNolintEnabled(true)
		domain.SetRequireSpecific(false)
	}()

	code := `package test
func foo() { //nolint
	defer func() {}()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err, "failed to parse code")

	// Find position on line 2
	var targetPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == 2 {
			targetPos = n.Pos()
			return false
		}
		return true
	})

	// When: Analyze code with bare //nolint comment (no analyzer name)
	// Then: Issue is NOT suppressed (bare nolint ignored)
	got := domain.ShouldSkip(fset, file.Comments, targetPos, "deferlint")
	assert.False(t, got, "ShouldSkip() should be false when require-specific is true and bare //nolint used")
}

func TestRequireSpecificAllowsSpecific(t *testing.T) {
	// Given: Config has nolint.require-specific: true
	domain.SetNolintEnabled(true)
	domain.SetRequireSpecific(true)
	defer func() {
		domain.SetNolintEnabled(true)
		domain.SetRequireSpecific(false)
	}()

	code := `package test
func foo() { //nolint:deferlint
	defer func() {}()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err, "failed to parse code")

	// Find position on line 2
	var targetPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == 2 {
			targetPos = n.Pos()
			return false
		}
		return true
	})

	// When: Analyze code with //nolint:deferlint comment
	// Then: Issue is suppressed (specific nolint honored)
	got := domain.ShouldSkip(fset, file.Comments, targetPos, "deferlint")
	assert.True(t, got, "ShouldSkip() should be true when require-specific is true and specific analyzer named")
}

func TestRequireSpecificWrongAnalyzer(t *testing.T) {
	// Given: Config has nolint.require-specific: true
	domain.SetNolintEnabled(true)
	domain.SetRequireSpecific(true)
	defer func() {
		domain.SetNolintEnabled(true)
		domain.SetRequireSpecific(false)
	}()

	code := `package test
func foo() { //nolint:errorlint
	defer func() {}()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err, "failed to parse code")

	// Find position on line 2
	var targetPos token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == 2 {
			targetPos = n.Pos()
			return false
		}
		return true
	})

	// When: Analyze deferlint issue with //nolint:errorlint comment
	// Then: Issue is NOT suppressed (wrong analyzer)
	got := domain.ShouldSkip(fset, file.Comments, targetPos, "deferlint")
	assert.False(t, got, "ShouldSkip() should be false when nolint specifies wrong analyzer")
}
