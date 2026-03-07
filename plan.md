# Go AI Code Analyzer - Implementation Plan

## Project Overview

A custom Go static analysis tool that detects common mistakes in AI-generated Go code. This linter will catch issues that existing tools miss, specifically targeting patterns that LLMs (ChatGPT, Copilot, Claude) frequently produce incorrectly.

## Project Context

**Repository**: github.com/curtbushko/go-ai-lint (existing repo)
**Architecture**: Hexagonal (Ports & Adapters)
**Methodology**: Test-Driven Development (TDD)
**Integration Target**: golangci-lint module plugin system

---

## Implementation TODO Checklist

This checklist follows strict TDD: write failing test FIRST, then implement, then refactor.

### Phase 1: Project Foundation

#### 1.1 Project Structure Setup
- [ ] Create directory structure following hexagonal architecture
  ```
  mkdir -p internal/core/{domain,ports,analyzers}
  mkdir -p internal/adapters/{analysis,reporters}
  mkdir -p cmd/go-ai-lint
  mkdir -p pkg/plugin
  ```
- [ ] Update go.mod with correct module name (`github.com/curtbushko/go-ai-lint`)
- [ ] Update .go-arch-lint.yml for new structure
- [ ] Create Makefile with targets: build, test, lint, arch-check

#### 1.2 Domain Layer (TDD)

**1.2.1 Severity Type**
- [ ] TEST (RED): Write `internal/core/domain/severity_test.go`
  - Test Severity constants exist (Critical, High, Medium, Low)
  - Test Severity.String() returns correct strings
  - Test Severity ordering (Critical > High > Medium > Low)
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/severity.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.2.2 Category Type**
- [ ] TEST (RED): Write `internal/core/domain/category_test.go`
  - Test Category constants exist (Defer, Context, Goroutine, Error, Nil, Type, Interface, Naming, Slice, String, Concurrency, Panic, Init, Option)
  - Test Category.String() returns correct strings
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/category.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.2.3 Position Type**
- [ ] TEST (RED): Write `internal/core/domain/position_test.go`
  - Test Position struct fields (Filename, Line, Column, EndLine, EndColumn)
  - Test Position.String() returns "file:line:col" format
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/position.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.2.4 FixExample Type**
- [ ] TEST (RED): Write `internal/core/domain/fix_example_test.go`
  - Test FixExample struct fields (Bad, Good, Explanation)
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/fix_example.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.2.5 Issue Type**
- [ ] TEST (RED): Write `internal/core/domain/issue_test.go`
  - Test Issue struct has all fields (ID, Name, Category, Severity, Position, Confidence, Message, Why, Fix, Example, CommonMistakes)
  - Test Issue.String() returns formatted message
  - Test NewIssue constructor
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/issue.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.2.6 DiagnosticTemplate Type**
- [ ] TEST (RED): Write `internal/core/domain/diagnostic_template_test.go`
  - Test DiagnosticTemplate struct for storing reusable diagnostic info
  - Test CreateIssue(position Position) creates Issue from template
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/diagnostic_template.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.2.7 Domain Errors**
- [ ] TEST (RED): Write `internal/core/domain/errors_test.go`
  - Test ErrInvalidSeverity, ErrInvalidCategory exist
  - Test error messages are descriptive
- [ ] RUN: `go test ./internal/core/domain/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/domain/errors.go`
- [ ] RUN: `go test ./internal/core/domain/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

#### 1.3 Ports Layer (TDD)

**1.3.1 Analyzer Port**
- [ ] TEST (RED): Write `internal/core/ports/analyzer_test.go`
  - Test Analyzer interface can be implemented (mock implementation)
  - Test interface has: Name(), Run(pass), BuildAnalyzer()
- [ ] RUN: `go test ./internal/core/ports/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/ports/analyzer.go`
- [ ] RUN: `go test ./internal/core/ports/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

**1.3.2 Reporter Port**
- [ ] TEST (RED): Write `internal/core/ports/reporter_test.go`
  - Test Reporter interface can be implemented
  - Test interface has: Report(issues []domain.Issue) error
  - Test Format constants (Text, JSON, AI, SARIF)
- [ ] RUN: `go test ./internal/core/ports/...` - confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/ports/reporter.go`
- [ ] RUN: `go test ./internal/core/ports/...` - confirm PASSES
- [ ] VALIDATE: `go build ./... && golangci-lint run`

#### 1.4 Phase 1 Validation
- [ ] Run full test suite: `go test -race -cover ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Run architecture check: `go-arch-lint check`
- [ ] Verify coverage >= 80% for domain and ports

---

### Phase 2: deferlint Analyzer (Sprint 1)

#### 2.1 AIL001: defer-in-loop

**2.1.1 Create Test Data**
- [ ] Create `testdata/src/deferlint/defer_in_loop.go` with:
  - Bad case: defer inside for loop with `// want "AIL001"`
  - Bad case: defer inside range loop with `// want "AIL001"`
  - Good case: defer in helper function (no want comment)
  - Good case: defer outside loop (no want comment)
  - Edge case: nested loops with defer

**2.1.2 Write Failing Test**
- [ ] TEST (RED): Write `internal/core/analyzers/deferlint/analyzer_test.go`
  - Use analysistest.Run with testdata
  - Test analyzer name is "deferlint"
  - Test analyzer doc string exists
- [ ] RUN: `go test ./internal/core/analyzers/deferlint/...` - confirm FAILS

**2.1.3 Implement Analyzer**
- [ ] IMPLEMENT (GREEN): Write `internal/core/analyzers/deferlint/analyzer.go`
  - Track loop depth (ForStmt, RangeStmt)
  - Detect DeferStmt when loopDepth > 0
  - Report with AIL001 message and AI-friendly guidance
- [ ] RUN: `go test ./internal/core/analyzers/deferlint/...` - confirm PASSES

**2.1.4 Validate**
- [ ] `go build ./...`
- [ ] `go test -race ./...`
- [ ] `golangci-lint run`
- [ ] `go-arch-lint check`

#### 2.2 AIL002: defer-close-error-ignored

**2.2.1 Create Test Data**
- [ ] Create `testdata/src/deferlint/defer_error_ignored.go` with:
  - Bad case: `defer file.Close()` without capturing error `// want "AIL002"`
  - Bad case: `defer resp.Body.Close()` `// want "AIL002"`
  - Good case: `defer func() { _ = file.Close() }()` (explicit ignore)
  - Good case: named return with deferred error check

**2.2.2 Write Failing Test**
- [ ] TEST (RED): Add test cases to `analyzer_test.go` for AIL002
- [ ] RUN: `go test ./internal/core/analyzers/deferlint/...` - confirm FAILS

**2.2.3 Implement Detection**
- [ ] IMPLEMENT (GREEN): Add AIL002 detection to analyzer.go
  - Find DeferStmt with CallExpr
  - Check if method returns error (use TypesInfo)
  - Check if error is captured
- [ ] RUN: `go test ./internal/core/analyzers/deferlint/...` - confirm PASSES

**2.2.4 Validate**
- [ ] Full validation suite

#### 2.3 AIL004: defer-nil-receiver

**2.3.1 Create Test Data**
- [ ] Create `testdata/src/deferlint/defer_nil_receiver.go` with:
  - Bad case: defer on var that could be nil `// want "AIL004"`
  - Good case: nil check before defer
  - Edge case: defer in error path

**2.3.2 Write Failing Test**
- [ ] TEST (RED): Add test cases to `analyzer_test.go` for AIL004
- [ ] RUN: confirm FAILS

**2.3.3 Implement Detection**
- [ ] IMPLEMENT (GREEN): Add AIL004 detection
- [ ] RUN: confirm PASSES

**2.3.4 Validate**
- [ ] Full validation suite

#### 2.4 Register deferlint with Plugin System
- [ ] Update plugin registration to include deferlint analyzer
- [ ] Write integration test for plugin loading
- [ ] Validate plugin works with golangci-lint

---

### Phase 3: contextlint Analyzer (Sprint 2)

#### 3.1 AIL010: context-todo-production

**3.1.1 Create Test Data**
- [ ] Create `testdata/src/contextlint/context_todo.go` with:
  - Bad case: `context.TODO()` in non-test file `// want "AIL010"`
  - Good case: `context.TODO()` in `*_test.go` file
  - Good case: `context.Background()` usage

**3.1.2 TDD Cycle**
- [ ] TEST (RED): Write `internal/core/analyzers/contextlint/analyzer_test.go`
- [ ] RUN: confirm FAILS
- [ ] IMPLEMENT (GREEN): Write `internal/core/analyzers/contextlint/analyzer.go`
- [ ] RUN: confirm PASSES
- [ ] VALIDATE: full suite

#### 3.2 AIL011: context-background-handler

**3.2.1 Create Test Data**
- [ ] Create `testdata/src/contextlint/context_handler.go` with:
  - Bad case: `context.Background()` in HTTP handler `// want "AIL011"`
  - Good case: `r.Context()` in handler

**3.2.2 TDD Cycle**
- [ ] TEST (RED): Add tests for AIL011
- [ ] RUN: confirm FAILS
- [ ] IMPLEMENT (GREEN): Detect context.Background() in handlers
- [ ] RUN: confirm PASSES
- [ ] VALIDATE: full suite

#### 3.3 AIL013: context-stored-in-struct

**3.3.1 Create Test Data**
- [ ] Create `testdata/src/contextlint/context_struct.go` with:
  - Bad case: struct with context.Context field `// want "AIL013"`
  - Good case: context passed as parameter

**3.3.2 TDD Cycle**
- [ ] TEST (RED): Add tests for AIL013
- [ ] RUN: confirm FAILS
- [ ] IMPLEMENT (GREEN): Detect context in struct fields
- [ ] RUN: confirm PASSES
- [ ] VALIDATE: full suite

---

### Phase 4: goroutinelint Analyzer (Sprint 3)

#### 4.1 AIL020: goroutine-no-cancel
- [ ] Create testdata with good/bad cases
- [ ] TEST (RED): Write analyzer test
- [ ] IMPLEMENT (GREEN): Detect goroutines without ctx.Done() check
- [ ] VALIDATE: full suite

#### 4.2 AIL021: goroutine-infinite-loop
- [ ] Create testdata with good/bad cases
- [ ] TEST (RED): Add tests
- [ ] IMPLEMENT (GREEN): Detect infinite loops without exit
- [ ] VALIDATE: full suite

#### 4.3 AIL022: goroutine-closure-capture (pre-1.22)
- [ ] Create testdata with good/bad cases
- [ ] TEST (RED): Add tests
- [ ] IMPLEMENT (GREEN): Detect loop var capture in goroutine
- [ ] VALIDATE: full suite

---

### Phase 5: slicemaplint Analyzer (Sprint 4)

#### 5.1 AIL060: nil-map-write
- [ ] Create testdata: `var m map[K]V; m[k] = v` `// want "AIL060"`
- [ ] TEST (RED): Write analyzer test
- [ ] IMPLEMENT (GREEN): Track map initialization, detect write to nil
- [ ] VALIDATE: full suite

#### 5.2 AIL061: slice-modify-during-iteration
- [ ] Create testdata with append/delete during range
- [ ] TEST (RED): Add tests
- [ ] IMPLEMENT (GREEN): Detect slice modification in range body
- [ ] VALIDATE: full suite

#### 5.3 AIL063: map-missing-comma-ok
- [ ] Create testdata with map access without comma-ok
- [ ] TEST (RED): Add tests
- [ ] IMPLEMENT (GREEN): Detect ambiguous map access
- [ ] VALIDATE: full suite

---

### Phase 6: errorlint Analyzer (Sprint 4)

#### 6.1 AIL030: error-handled-twice
- [ ] Create testdata: log error AND return it `// want "AIL030"`
- [ ] TEST (RED): Write analyzer test
- [ ] IMPLEMENT (GREEN): Detect log+return pattern
- [ ] VALIDATE: full suite

#### 6.2 AIL033: error-fmt-not-wrapped
- [ ] Create testdata: `fmt.Errorf("...: %v", err)` `// want "AIL033"`
- [ ] TEST (RED): Add tests
- [ ] IMPLEMENT (GREEN): Detect %v instead of %w
- [ ] VALIDATE: full suite

---

### Phase 7: concurrencylint Analyzer (Sprint 6)

#### 7.1 AIL080: waitgroup-done-not-deferred
- [ ] Create testdata: `wg.Done()` not in defer `// want "AIL080"`
- [ ] TEST (RED): Write analyzer test
- [ ] IMPLEMENT (GREEN): Detect wg.Done() outside defer
- [ ] VALIDATE: full suite

#### 7.2 AIL082: select-only-default
- [ ] Create testdata: select with only default case
- [ ] TEST (RED): Add tests
- [ ] IMPLEMENT (GREEN): Detect useless select
- [ ] VALIDATE: full suite

---

### Phase 8: Remaining Analyzers (Sprints 5-8)

#### 8.1 naminglint
- [ ] AIL050: getter-with-get-prefix - TDD cycle
- [ ] AIL051: redundant-package-prefix - TDD cycle

#### 8.2 interfacelint
- [ ] AIL040: interface-too-large - TDD cycle
- [ ] AIL042: interface-missing-er-suffix - TDD cycle

#### 8.3 stringlint
- [ ] AIL070: string-byte-iteration - TDD cycle
- [ ] AIL071: string-concat-in-loop - TDD cycle

#### 8.4 paniclint
- [ ] AIL090: panic-in-library - TDD cycle
- [ ] AIL091: empty-recover - TDD cycle

#### 8.5 initlint
- [ ] AIL100: init-with-network - TDD cycle
- [ ] AIL101: init-with-file-io - TDD cycle

#### 8.6 optionlint
- [ ] AIL110: with-not-option - TDD cycle

---

### Phase 9: Adapters Layer

#### 9.1 Analysis Adapter
- [ ] TEST (RED): Write `internal/adapters/analysis/adapter_test.go`
  - Test wrapping domain analyzers for go/analysis
- [ ] IMPLEMENT (GREEN): Write adapter
- [ ] VALIDATE: full suite

#### 9.2 Text Reporter
- [ ] TEST (RED): Write `internal/adapters/reporters/text_test.go`
- [ ] IMPLEMENT (GREEN): Write text reporter
- [ ] VALIDATE: full suite

#### 9.3 JSON Reporter
- [ ] TEST (RED): Write `internal/adapters/reporters/json_test.go`
- [ ] IMPLEMENT (GREEN): Write JSON reporter
- [ ] VALIDATE: full suite

#### 9.4 AI Reporter (AI-friendly output)
- [ ] TEST (RED): Write `internal/adapters/reporters/ai_test.go`
- [ ] IMPLEMENT (GREEN): Write AI reporter with guidance
- [ ] VALIDATE: full suite

---

### Phase 10: CLI and Plugin Integration

#### 10.1 CLI Entry Point
- [ ] Create `cmd/go-ai-lint/main.go` using multichecker
- [ ] TEST: Manual testing with sample code
- [ ] VALIDATE: `go build ./cmd/go-ai-lint`

#### 10.2 golangci-lint Plugin
- [ ] Update `pkg/plugin/plugin.go` with all analyzers
- [ ] TEST: Integration test with golangci-lint
- [ ] Document plugin configuration

#### 10.3 Configuration
- [ ] Implement config loading from `.go-ai-lint.yml`
- [ ] TEST: Config parsing tests
- [ ] Document configuration options

---

### Phase 11: Final Validation

#### 11.1 Quality Gates
- [ ] All tests pass: `go test -race -cover ./...`
- [ ] Coverage >= 80%: `go test -coverprofile=coverage.out ./...`
- [ ] Lint clean: `golangci-lint run`
- [ ] Architecture clean: `go-arch-lint check`
- [ ] No false positives on Go stdlib sample

#### 11.2 Documentation
- [ ] README.md with usage instructions
- [ ] Document each analyzer with examples
- [ ] Document golangci-lint integration

#### 11.3 Release
- [ ] Tag v1.0.0
- [ ] Create GitHub release
- [ ] Publish to pkg.go.dev

---

## Problem Statement

AI code generators consistently produce Go code with predictable mistakes:

| Problem | Severity | Existing Coverage | Our Focus |
|---------|----------|-------------------|-----------|
| Defer in loops | Critical | Partial (revive) | **Primary** |
| Ignored defer errors | Critical | None | **Primary** |
| Nil map write | Critical | Partial (staticcheck) | **Primary** |
| context.TODO() in production | High | None | **Primary** |
| Goroutines without cancellation | High | None | **Primary** |
| WaitGroup.Done() not deferred | High | None | **Primary** |
| Panic in library code | High | None | **Primary** |
| Unsafe type assertions | High | forcetypeassert | Secondary |
| context.Background() misuse | Medium | Partial | **Primary** |
| Error handling twice | Medium | None | **Primary** |
| GetX() instead of X() | Medium | None | **Primary** |
| God interfaces (>5 methods) | Medium | interfacebloat | **Primary** |
| Slice modification in loop | Medium | None | **Primary** |
| String concat in loop | Medium | gocritic | **Primary** |
| Init with side effects | Medium | None | **Primary** |
| Missing comma-ok on map | Medium | Partial | Secondary |

---

## Architecture: Hexagonal Design

```
go-ai-lint/
├── cmd/                          # Application entry points
│   └── go-ai-lint/
│       └── main.go               # CLI using singlechecker/multichecker
│
├── internal/
│   ├── core/                     # INNER LAYERS (Business Logic)
│   │   ├── domain/               # Domain types and errors
│   │   │   ├── issue.go          # Issue represents a detected problem
│   │   │   ├── severity.go       # Severity levels (Critical, High, Medium)
│   │   │   ├── category.go       # Issue categories (Defer, Context, etc.)
│   │   │   └── errors.go         # Domain errors
│   │   │
│   │   ├── ports/                # Interfaces (contracts)
│   │   │   ├── analyzer.go       # Analyzer port interface
│   │   │   ├── reporter.go       # Reporter port interface
│   │   │   └── detector.go       # Pattern detector interface
│   │   │
│   │   └── analyzers/            # Use cases - individual analyzers
│   │       ├── deferlint/        # Defer mistake analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── contextlint/      # Context misuse analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── goroutinelint/    # Goroutine lifecycle analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── errorlint/        # Error handling analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── interfacelint/    # Interface design analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── naminglint/       # Go naming convention analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── slicemaplint/     # Slice/map pitfall analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── stringlint/       # String handling analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── concurrencylint/  # Additional concurrency analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── paniclint/        # Panic/recover analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       ├── initlint/         # Init function analyzer
│   │       │   ├── analyzer.go
│   │       │   └── analyzer_test.go
│   │       └── optionlint/       # Functional options analyzer
│   │           ├── analyzer.go
│   │           └── analyzer_test.go
│   │
│   ├── adapters/                 # OUTER LAYER (Infrastructure)
│   │   ├── analysis/             # go/analysis framework adapter
│   │   │   ├── adapter.go        # Wraps analyzers for go/analysis
│   │   │   └── inspector.go      # AST inspection utilities
│   │   │
│   │   └── reporters/            # Output adapters
│   │       ├── text.go           # Text output
│   │       ├── json.go           # JSON output
│   │       └── sarif.go          # SARIF format (for IDE integration)
│   │
│   └── config/                   # Configuration loading
│       └── config.go
│
├── pkg/                          # Public API (for golangci-lint plugin)
│   └── plugin/
│       └── plugin.go             # Exposes New() for golangci-lint
│
├── testdata/                     # Test fixtures
│   └── src/
│       ├── deferlint/
│       ├── contextlint/
│       ├── goroutinelint/
│       ├── errorlint/
│       ├── interfacelint/
│       ├── naminglint/
│       ├── slicemaplint/
│       ├── stringlint/
│       ├── concurrencylint/
│       ├── paniclint/
│       ├── initlint/
│       └── optionlint/
│
├── .golangci.yml                 # Our own linting config
├── .go-arch-lint.yml             # Architecture enforcement
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Architecture Rules (.go-arch-lint.yml)

```yaml
version: 3
workdir: internal

components:
  domain:
    in: core/domain/**
  ports:
    in: core/ports/**
  analyzers:
    in: core/analyzers/**
  # Individual analyzer components for fine-grained control
  deferlint:
    in: core/analyzers/deferlint/**
  contextlint:
    in: core/analyzers/contextlint/**
  goroutinelint:
    in: core/analyzers/goroutinelint/**
  errorlint:
    in: core/analyzers/errorlint/**
  interfacelint:
    in: core/analyzers/interfacelint/**
  naminglint:
    in: core/analyzers/naminglint/**
  slicemaplint:
    in: core/analyzers/slicemaplint/**
  stringlint:
    in: core/analyzers/stringlint/**
  concurrencylint:
    in: core/analyzers/concurrencylint/**
  paniclint:
    in: core/analyzers/paniclint/**
  initlint:
    in: core/analyzers/initlint/**
  optionlint:
    in: core/analyzers/optionlint/**
  adapters:
    in: adapters/**
  config:
    in: config/**

commonComponents:
  - domain
  - ports

deps:
  analyzers:
    mayDependOn:
      - ports
      - domain
  adapters:
    mayDependOn:
      - ports
      - domain
  domain:
    # No dependencies - pure domain
  ports:
    mayDependOn:
      - domain
  config:
    canUse:
      - os
      - encoding/json
```

---

## Domain Model

### Issue (domain/issue.go)

```go
// Issue represents a detected code problem.
// Designed for both human and AI consumption.
type Issue struct {
    ID          string      // Unique identifier (e.g., "AIL001")
    Name        string      // Short name (e.g., "defer-in-loop")
    Category    Category    // Category (Defer, Context, Goroutine, Error)
    Severity    Severity    // Severity level
    Position    Position    // File location
    Confidence  float64     // 0.0-1.0 confidence level

    // Human-readable
    Message     string      // What was detected

    // AI-consumable guidance (prevents fix loops)
    Why              string      // Why this is a problem (consequences)
    Fix              string      // How to fix it (strategy)
    Example          FixExample  // Before/after code
    CommonMistakes   []string    // What NOT to do when fixing
}

// FixExample provides concrete before/after code.
type FixExample struct {
    Bad         string  // The problematic pattern
    Good        string  // The correct pattern
    Explanation string  // Why the good version works
}

// Position represents a location in source code.
type Position struct {
    Filename string
    Line     int
    Column   int
    EndLine  int
    EndCol   int
}
```

### Severity (domain/severity.go)

```go
type Severity int

const (
    SeverityCritical Severity = iota // Will likely cause bugs/panics
    SeverityHigh                      // Likely problematic
    SeverityMedium                    // Code smell, should fix
    SeverityLow                       // Suggestion
)
```

### Category (domain/category.go)

```go
type Category string

const (
    CategoryDefer     Category = "defer"
    CategoryContext   Category = "context"
    CategoryGoroutine Category = "goroutine"
    CategoryError     Category = "error"
    CategoryNil       Category = "nil"
    CategoryType      Category = "type"
)
```

---

## AI-Friendly Output Design

### Problem: AI Fix Loops

When AI encounters linting errors, it can get stuck in loops:
1. Sees error "defer in loop"
2. Removes defer entirely (wrong fix)
3. Sees new error "file handle not closed"
4. Adds defer back in loop (original problem)
5. Repeat forever

### Solution: Rich Diagnostic Messages

Every diagnostic MUST include:

1. **What** - The specific problem detected
2. **Why** - The consequence (why it matters)
3. **How** - Concrete fix strategy
4. **Example** - Before/after code
5. **Pitfalls** - Common mistakes when fixing

### Output Formats

#### Standard (Human)
```
service.go:42:3: AIL001 defer-in-loop: defer inside loop delays resource cleanup until function returns
```

#### AI Mode (--format=ai)
```json
{
  "id": "AIL001",
  "name": "defer-in-loop",
  "file": "service.go",
  "line": 42,
  "column": 3,
  "severity": "critical",
  "message": "defer inside loop delays resource cleanup until function returns",
  "why": "Deferred calls accumulate until the function returns. In a loop processing N items, all N resources stay open simultaneously, risking resource exhaustion (file descriptors, memory, connections).",
  "fix": "Extract the loop body into a separate function. The defer will then execute after each iteration when the helper function returns.",
  "example": {
    "bad": "for _, f := range files {\n    file, _ := os.Open(f)\n    defer file.Close()  // All files stay open!\n    process(file)\n}",
    "good": "for _, f := range files {\n    if err := processFile(f); err != nil {\n        return err\n    }\n}\n\nfunc processFile(path string) error {\n    file, err := os.Open(path)\n    if err != nil {\n        return err\n    }\n    defer file.Close()  // Closes after each file\n    return process(file)\n}",
    "explanation": "By extracting to a helper function, defer runs when that function returns (after each iteration), not when the outer function returns."
  },
  "common_mistakes": [
    "WRONG: Removing defer entirely - resource will never be closed",
    "WRONG: Using a closure `func() { defer f.Close() }()` - works but adds overhead",
    "WRONG: Manually calling Close() without defer - may skip on early return/panic",
    "WRONG: Moving defer outside the loop - only closes the last resource"
  ]
}
```

### Reporter Interface (ports/reporter.go)

```go
// Reporter outputs issues in various formats.
type Reporter interface {
    Report(issues []domain.Issue) error
}

// Format specifies output format.
type Format string

const (
    FormatText   Format = "text"    // Human-readable
    FormatJSON   Format = "json"    // Machine-parseable
    FormatAI     Format = "ai"      // AI-optimized with guidance
    FormatSARIF  Format = "sarif"   // IDE integration
)
```

### Diagnostic Registry

Each analyzer registers its diagnostics with full guidance:

```go
// internal/core/analyzers/deferlint/diagnostics.go
package deferlint

import "github.com/curtbushko/go-ai-lint/internal/core/domain"

var Diagnostics = map[string]domain.DiagnosticTemplate{
    "AIL001": {
        ID:       "AIL001",
        Name:     "defer-in-loop",
        Severity: domain.SeverityCritical,
        Category: domain.CategoryDefer,
        Message:  "defer inside loop delays resource cleanup until function returns",
        Why: `Deferred calls accumulate until the function returns. In a loop
processing N items, all N resources stay open simultaneously, risking
resource exhaustion (file descriptors, memory, database connections).

For example, processing 10,000 files would open 10,000 file handles
before closing any of them.`,
        Fix: `Extract the loop body into a separate function. The defer will
then execute after each iteration when the helper function returns.

Alternative: Use an immediately-invoked function literal (closure),
though this adds slight overhead.`,
        Example: domain.FixExample{
            Bad: `for _, f := range files {
    file, _ := os.Open(f)
    defer file.Close()  // All files stay open!
    process(file)
}`,
            Good: `for _, f := range files {
    if err := processFile(f); err != nil {
        return err
    }
}

func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()  // Closes after each file
    return process(file)
}`,
            Explanation: "By extracting to a helper function, defer runs when that function returns (after each iteration), not when the outer function returns.",
        },
        CommonMistakes: []string{
            "WRONG: Removing defer entirely - resource will never be closed on early return or panic",
            "WRONG: Moving defer outside the loop - only closes the last resource assigned to the variable",
            "WRONG: Manually calling Close() without defer - may skip on early return or panic",
        },
    },
    // ... more diagnostics
}
```

### Integration with Claude Code

Add a hook or skill instruction to use AI format:

```bash
# In quality-gates.sh or as a Claude hook
go-ai-lint --format=ai ./... 2>&1 | jq -r '.[] | "
ERROR: \(.id) \(.name)
FILE: \(.file):\(.line)
PROBLEM: \(.message)
WHY: \(.why)
FIX: \(.fix)
EXAMPLE (bad):
\(.example.bad)

EXAMPLE (good):
\(.example.good)

DO NOT:
\(.common_mistakes | join("\n"))
"'
```

---

## Analyzer Specifications

### 1. deferlint - Defer Mistake Analyzer

**ID Prefix**: AIL001-AIL009

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL001 | defer-in-loop | Critical | Defer inside for/range loop |
| AIL002 | defer-close-error-ignored | High | Deferred Close() error ignored |
| AIL003 | defer-flush-error-ignored | High | Deferred Flush() error ignored |
| AIL004 | defer-nil-receiver | Critical | Defer on potentially nil receiver |
| AIL005 | defer-arg-evaluation | Medium | Defer argument evaluated immediately |

**Detection Logic**:

```
AIL001: defer-in-loop
├── Track loop depth (ForStmt, RangeStmt enter/exit)
├── On DeferStmt, check if loopDepth > 0
└── Report with suggestion to wrap in closure

AIL002/003: defer-close-error-ignored
├── Find DeferStmt with CallExpr
├── Check if method is Close/Flush/Sync
├── Check if method returns error (via TypesInfo)
└── Report if error not captured

AIL004: defer-nil-receiver
├── Find DeferStmt with method call
├── Trace receiver back to assignment
├── Check if receiver could be nil (error path)
└── Report if nil possible

AIL005: defer-arg-evaluation
├── Find DeferStmt with arguments
├── Check if arguments have side effects (time.Now(), etc.)
├── Suggest wrapping in closure if needed
└── Report common patterns (time.Since, fmt.Sprintf with changing vars)
```

### 2. contextlint - Context Misuse Analyzer

**ID Prefix**: AIL010-AIL019

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL010 | context-todo-production | High | context.TODO() in non-test code |
| AIL011 | context-background-handler | Medium | context.Background() in HTTP handler |
| AIL012 | context-not-first-param | Low | Context not first parameter |
| AIL013 | context-stored-in-struct | Medium | Context stored in struct field |

**Detection Logic**:

```
AIL010: context-todo-production
├── Find CallExpr for context.TODO()
├── Check if file is *_test.go
├── If not test file, report
└── Suggest using passed context or context.Background()

AIL011: context-background-handler
├── Find functions with http.ResponseWriter, *http.Request params
├── Find context.Background() calls inside
├── Report - should use r.Context()
└── Suggest: ctx := r.Context()

AIL012: context-not-first-param
├── Find FuncDecl with context.Context parameter
├── Check if it's the first parameter
└── Report if not first (per Go convention)

AIL013: context-stored-in-struct
├── Find StructType with context.Context field
├── Report - contexts should be passed, not stored
└── Except: allow if field name suggests it's for cancellation
```

### 3. goroutinelint - Goroutine Lifecycle Analyzer

**ID Prefix**: AIL020-AIL029

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL020 | goroutine-no-cancel | High | Goroutine with no cancellation mechanism |
| AIL021 | goroutine-infinite-loop | Critical | Infinite loop in goroutine without context |
| AIL022 | goroutine-closure-capture | Medium | Loop variable captured in goroutine (pre-1.22) |
| AIL023 | goroutine-fire-and-forget | Medium | Goroutine started with no wait/sync |

**Detection Logic**:

```
AIL020: goroutine-no-cancel
├── Find GoStmt (go keyword)
├── Analyze closure/function body
├── Check for: ctx.Done(), select with done channel, return path
├── Report if no cancellation mechanism found
└── Suggest: add context parameter with ctx.Done() check

AIL021: goroutine-infinite-loop
├── Find GoStmt containing ForStmt with no condition
├── Or: for { ... } pattern
├── Check if there's a return/break with context
├── Report if truly infinite with no exit
└── Severity: Critical (guaranteed goroutine leak)

AIL022: goroutine-closure-capture
├── Check Go version (skip if >= 1.22)
├── Find GoStmt inside RangeStmt/ForStmt
├── Check if loop variable used in closure
├── Report if captured without copy
└── Suggest: add `v := v` or use explicit parameter

AIL023: goroutine-fire-and-forget
├── Find GoStmt
├── Check enclosing function for WaitGroup, channel, sync
├── Report if goroutine has no synchronization
└── Lower severity - sometimes intentional
```

### 4. errorlint - Error Handling Analyzer

**ID Prefix**: AIL030-AIL039

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL030 | error-handled-twice | Medium | Error both logged and returned |
| AIL031 | error-nil-check-missing | High | Function returns (T, error) but only T used |
| AIL032 | error-shadow-in-block | Medium | err shadowed in nested block |
| AIL033 | error-fmt-not-wrapped | Medium | Error returned with fmt.Errorf but no %w |

**Detection Logic**:

```
AIL030: error-handled-twice
├── Find IfStmt checking err != nil
├── Check if block contains log.* or fmt.* call with err
├── Check if block also returns err
├── Report - error should be logged OR returned, not both
└── Suggest: remove logging or wrap and return

AIL031: error-nil-check-missing
├── Find AssignStmt with CallExpr on RHS
├── Check if function returns (..., error)
├── Check if error is assigned to _
├── Report - error should be checked
└── Note: errcheck covers this, but we add AI-specific suggestions

AIL032: error-shadow-in-block
├── Track err variable scope
├── Find := assignment that shadows outer err
├── Report if outer err was unchecked
└── Suggest: use = instead of := or rename

AIL033: error-fmt-not-wrapped
├── Find fmt.Errorf calls
├── Check if first arg contains error but no %w
├── Report - should use %w for error wrapping
└── Suggest: change %v to %w
```

### 5. interfacelint - Interface Design Analyzer

**ID Prefix**: AIL040-AIL049

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL040 | interface-too-large | Medium | Interface has > 5 methods |
| AIL041 | interface-wrong-location | Medium | Interface defined at implementation site |
| AIL042 | interface-missing-er-suffix | Low | Single-method interface not named with "-er" |
| AIL043 | interface-accepts-concrete | Medium | Function accepts concrete type where interface suffices |

**Detection Logic**:

```
AIL040: interface-too-large
├── Find InterfaceType in TypeSpec
├── Count methods in interface
├── Report if count > 5
└── Suggest: split into smaller, focused interfaces

AIL041: interface-wrong-location
├── Find InterfaceType definitions
├── Check if same package has implementation
├── Report if interface is in implementer's package
└── Suggest: define interfaces where they're used

AIL042: interface-missing-er-suffix
├── Find InterfaceType with exactly 1 method
├── Check if interface name ends with "-er"
├── Report if single-method interface lacks "-er" suffix
└── Suggest: rename Reader, Writer, Closer, etc.

AIL043: interface-accepts-concrete
├── Find FuncDecl with pointer parameters
├── Check if only 1-2 methods are called on param
├── Report if small interface would suffice
└── Suggest: accept interface instead of concrete type
```

### 6. naminglint - Go Naming Convention Analyzer

**ID Prefix**: AIL050-AIL059

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL050 | getter-with-get-prefix | Medium | `GetX()` should be `X()` |
| AIL051 | redundant-package-prefix | Medium | `user.UserService` redundancy |
| AIL052 | package-name-underscores | Low | Package name contains underscore |
| AIL053 | exported-in-internal | Low | Exported type in internal package |

**Detection Logic**:

```
AIL050: getter-with-get-prefix
├── Find FuncDecl starting with "Get"
├── Check if function has no parameters (or just ctx)
├── Check if function returns single value
├── Report - should use X() not GetX()
└── Suggest: rename GetUser() to User()

AIL051: redundant-package-prefix
├── Find FuncDecl or TypeSpec
├── Check if name starts with package name
├── Report redundancy
└── Suggest: user.Service not user.UserService

AIL052: package-name-underscores
├── Check package declaration
├── Report if contains underscore or mixed case
└── Suggest: use lowercase single word

AIL053: exported-in-internal
├── Check if file path contains "/internal/"
├── Find exported (uppercase) declarations
├── Report - internal packages shouldn't export widely
└── Note: Lower severity, sometimes intentional
```

### 7. slicemaplint - Slice/Map Pitfall Analyzer

**ID Prefix**: AIL060-AIL069

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL060 | nil-map-write | Critical | Writing to nil map (panic) |
| AIL061 | slice-modify-during-iteration | High | Modifying slice in range loop |
| AIL062 | slice-append-side-effect | Medium | append() may modify underlying array |
| AIL063 | map-missing-comma-ok | Medium | Map access without comma-ok idiom |

**Detection Logic**:

```
AIL060: nil-map-write
├── Find IndexExpr on LHS of assignment (m[key] = value)
├── Check if map variable was declared but not initialized
├── Trace back to var declaration without make()
├── Report - will panic at runtime
└── Suggest: use make(map[K]V) before write

AIL061: slice-modify-during-iteration
├── Find RangeStmt over slice
├── Check body for append/delete/slice reassignment
├── Report - corrupts iteration
└── Suggest: iterate backwards or collect indices first

AIL062: slice-append-side-effect
├── Find slice assignment b := a[:n]
├── Find subsequent append(b, ...)
├── Check if a is used after append
├── Report - may modify original slice
└── Suggest: use copy() or full slice expression a[:n:n]

AIL063: map-missing-comma-ok
├── Find IndexExpr on map type
├── Check if result used without comma-ok
├── Check if zero value could be valid
├── Report if ambiguous (e.g., map[string]int where 0 is valid)
└── Suggest: use value, ok := m[key]
```

### 8. stringlint - String Handling Analyzer

**ID Prefix**: AIL070-AIL079

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL070 | string-byte-iteration | Medium | Byte iteration on UTF-8 string |
| AIL071 | string-concat-in-loop | Medium | O(n²) string concatenation |
| AIL072 | strings-contains-vs-index | Low | Verbose index check vs Contains |

**Detection Logic**:

```
AIL070: string-byte-iteration
├── Find ForStmt with i := 0; i < len(s); i++
├── Check if s is string type
├── Check if body accesses s[i]
├── Report - breaks on multi-byte UTF-8
└── Suggest: use for _, r := range s

AIL071: string-concat-in-loop
├── Find ForStmt or RangeStmt
├── Find AssignStmt with += on string
├── Report - O(n²) complexity
└── Suggest: use strings.Builder

AIL072: strings-contains-vs-index
├── Find BinaryExpr: strings.Index(s, x) != -1
├── Or: strings.Index(s, x) >= 0
├── Report - verbose pattern
└── Suggest: use strings.Contains(s, x)
```

### 9. concurrencylint - Additional Concurrency Analyzer

**ID Prefix**: AIL080-AIL089

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL080 | waitgroup-done-not-deferred | High | wg.Done() not in defer |
| AIL081 | channel-never-closed | Medium | Channel created but never closed |
| AIL082 | select-only-default | Medium | select with only default case |
| AIL083 | unbuffered-send-in-func | Medium | Sending to unbuffered channel without goroutine |

**Detection Logic**:

```
AIL080: waitgroup-done-not-deferred
├── Find CallExpr for wg.Done()
├── Check if it's inside DeferStmt
├── If not in defer, check if all paths call Done
├── Report - may leak if function panics/returns early
└── Suggest: defer wg.Done() at start of goroutine

AIL081: channel-never-closed
├── Find make(chan T) assignments
├── Track if close(ch) ever called in same scope
├── Report if channel never closed
└── Note: Lower confidence, may be intentional

AIL082: select-only-default
├── Find SelectStmt
├── Check if only has default case
├── Report - select with only default is useless
└── Suggest: remove select entirely

AIL083: unbuffered-send-in-func
├── Find make(chan T) with no capacity
├── Find send operation ch <- v in same function
├── Report if no goroutine to receive
└── Suggest: use buffered channel or separate goroutine
```

### 10. paniclint - Panic/Recover Analyzer

**ID Prefix**: AIL090-AIL099

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL090 | panic-in-library | High | panic() in non-main package |
| AIL091 | empty-recover | Medium | recover() result not used |
| AIL092 | must-pattern-undoc | Low | MustX() without panic documentation |

**Detection Logic**:

```
AIL090: panic-in-library
├── Check if package is "main"
├── Find CallExpr for panic()
├── Skip if function name starts with "Must"
├── Report - libraries should return errors
└── Suggest: return error instead of panic

AIL091: empty-recover
├── Find DeferStmt with recover()
├── Check if recover() return value is used
├── Report if result discarded
└── Suggest: log or re-panic with context

AIL092: must-pattern-undoc
├── Find FuncDecl starting with "Must"
├── Check if doc comment mentions panic
├── Report if panic undocumented
└── Suggest: document when function panics
```

### 11. initlint - Init Function Analyzer

**ID Prefix**: AIL100-AIL109

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL100 | init-with-network | High | Network calls in init() |
| AIL101 | init-with-file-io | Medium | File I/O in init() |
| AIL102 | init-too-complex | Medium | init() with > 10 statements |

**Detection Logic**:

```
AIL100: init-with-network
├── Find FuncDecl named "init"
├── Search body for http.*, net.*, grpc.* calls
├── Report - init should be deterministic
└── Suggest: move to explicit initialization function

AIL101: init-with-file-io
├── Find FuncDecl named "init"
├── Search body for os.Open, os.Create, ioutil.Read*
├── Report - init should avoid side effects
└── Suggest: use sync.Once or explicit init

AIL102: init-too-complex
├── Find FuncDecl named "init"
├── Count statements in body
├── Report if > 10 statements
└── Suggest: extract to named init function
```

### 12. optionlint - Functional Options Analyzer

**ID Prefix**: AIL110-AIL119

| ID | Name | Severity | Description |
|----|------|----------|-------------|
| AIL110 | with-not-option | Medium | WithX() doesn't follow option pattern |
| AIL111 | option-modifies-struct | Medium | Option modifies struct instead of config |

**Detection Logic**:

```
AIL110: with-not-option
├── Find FuncDecl starting with "With"
├── Check if return type is func(*Config) or similar
├── Report if With* doesn't return option function
└── Suggest: follow functional options pattern

AIL111: option-modifies-struct
├── Find option function (returns func(*T))
├── Check if it modifies exported struct fields directly
├── Report - options should modify config, not final struct
└── Suggest: use internal config struct
```

---

## TDD Implementation Plan

For each analyzer, follow this workflow:

### Phase 1: INVESTIGATE (Per Analyzer)

- [ ] Research the specific problem pattern
- [ ] Find real-world examples in AI-generated code
- [ ] Document edge cases and false positive risks
- [ ] Define acceptance criteria

### Phase 2: PLAN (Per Analyzer)

- [ ] Design the detection algorithm
- [ ] Identify AST nodes to inspect
- [ ] Determine if TypesInfo needed
- [ ] Plan test cases (positive and negative)

### Phase 3: TEST (RED)

- [ ] Create testdata files with `// want` comments
- [ ] Write test using analysistest.Run
- [ ] Run test - confirm it FAILS
- [ ] Commit failing test

### Phase 4: IMPLEMENT (GREEN)

- [ ] Implement minimum code to pass test
- [ ] Run test - confirm it PASSES
- [ ] Add more test cases incrementally
- [ ] Commit passing implementation

### Phase 5: VALIDATE

- [ ] Run all tests: `go test ./...`
- [ ] Run with race detector: `go test -race ./...`
- [ ] Run golangci-lint: `golangci-lint run`
- [ ] Run architecture check: `go-arch-lint check`
- [ ] Check coverage: `go test -cover ./...`

### Phase 6: REFACTOR

- [ ] Clean up code
- [ ] Extract common utilities
- [ ] Improve error messages
- [ ] Add documentation
- [ ] Commit refactored code

---

## Implementation Order

Prioritized by value and complexity:

### Sprint 1: Foundation + deferlint

1. **Project Setup** (Day 1)
   - [ ] Initialize go.mod
   - [ ] Create directory structure
   - [ ] Add .golangci.yml and .go-arch-lint.yml
   - [ ] Create Makefile
   - [ ] Set up CI (GitHub Actions)

2. **Domain Layer** (Day 1-2)
   - [ ] Define Issue type
   - [ ] Define Severity enum
   - [ ] Define Category enum
   - [ ] Define Position type
   - [ ] Write tests for domain types

3. **Ports Layer** (Day 2)
   - [ ] Define Analyzer interface
   - [ ] Define Detector interface
   - [ ] Define Reporter interface

4. **deferlint Analyzer** (Day 3-5)
   - [ ] AIL001: defer-in-loop (TDD)
   - [ ] AIL002: defer-close-error-ignored (TDD)
   - [ ] AIL004: defer-nil-receiver (TDD)

5. **Adapters** (Day 5-6)
   - [ ] go/analysis adapter
   - [ ] Text reporter
   - [ ] CLI entry point

6. **Integration** (Day 6-7)
   - [ ] multichecker main.go
   - [ ] golangci-lint plugin
   - [ ] Documentation

### Sprint 2: contextlint

1. **contextlint Analyzer** (Day 1-4)
   - [ ] AIL010: context-todo-production (TDD)
   - [ ] AIL011: context-background-handler (TDD)
   - [ ] AIL013: context-stored-in-struct (TDD)

2. **Integration** (Day 5)
   - [ ] Add to multichecker
   - [ ] Update plugin
   - [ ] Update documentation

### Sprint 3: goroutinelint

1. **goroutinelint Analyzer** (Day 1-5)
   - [ ] AIL020: goroutine-no-cancel (TDD)
   - [ ] AIL021: goroutine-infinite-loop (TDD)
   - [ ] AIL022: goroutine-closure-capture (TDD)

2. **Integration** (Day 6)
   - [ ] Add to multichecker
   - [ ] Update plugin

### Sprint 4: errorlint + slicemaplint

1. **errorlint Analyzer** (Day 1-3)
   - [ ] AIL030: error-handled-twice (TDD)
   - [ ] AIL033: error-fmt-not-wrapped (TDD)

2. **slicemaplint Analyzer** (Day 4-6)
   - [ ] AIL060: nil-map-write (TDD)
   - [ ] AIL061: slice-modify-during-iteration (TDD)
   - [ ] AIL063: map-missing-comma-ok (TDD)

### Sprint 5: naminglint + interfacelint

1. **naminglint Analyzer** (Day 1-3)
   - [ ] AIL050: getter-with-get-prefix (TDD)
   - [ ] AIL051: redundant-package-prefix (TDD)

2. **interfacelint Analyzer** (Day 4-6)
   - [ ] AIL040: interface-too-large (TDD)
   - [ ] AIL042: interface-missing-er-suffix (TDD)

### Sprint 6: stringlint + concurrencylint

1. **stringlint Analyzer** (Day 1-3)
   - [ ] AIL070: string-byte-iteration (TDD)
   - [ ] AIL071: string-concat-in-loop (TDD)

2. **concurrencylint Analyzer** (Day 4-6)
   - [ ] AIL080: waitgroup-done-not-deferred (TDD)
   - [ ] AIL082: select-only-default (TDD)

### Sprint 7: paniclint + initlint

1. **paniclint Analyzer** (Day 1-3)
   - [ ] AIL090: panic-in-library (TDD)
   - [ ] AIL091: empty-recover (TDD)

2. **initlint Analyzer** (Day 4-6)
   - [ ] AIL100: init-with-network (TDD)
   - [ ] AIL101: init-with-file-io (TDD)

### Sprint 8: optionlint + Polish

1. **optionlint Analyzer** (Day 1-2)
   - [ ] AIL110: with-not-option (TDD)

2. **Polish** (Day 3-5)
   - [ ] Performance optimization
   - [ ] False positive reduction
   - [ ] Documentation
   - [ ] Release v1.0.0

---

## Test Data Examples

### deferlint/testdata/src/deferlint/defer_in_loop.go

```go
package deferlint

import "os"

func BadDeferInLoop(files []string) error {
    for _, f := range files {
        file, err := os.Open(f)
        if err != nil {
            return err
        }
        defer file.Close() // want "AIL001: defer inside loop"
    }
    return nil
}

func GoodDeferWithClosure(files []string) error {
    for _, f := range files {
        if err := processFile(f); err != nil {
            return err
        }
    }
    return nil
}

func processFile(name string) error {
    file, err := os.Open(name)
    if err != nil {
        return err
    }
    defer file.Close() // OK - not in loop
    return nil
}
```

### contextlint/testdata/src/contextlint/context_todo.go

```go
package contextlint

import (
    "context"
    "database/sql"
)

func BadContextTODO(db *sql.DB) error {
    ctx := context.TODO() // want "AIL010: context.TODO"
    return db.PingContext(ctx)
}

func GoodContextParam(ctx context.Context, db *sql.DB) error {
    return db.PingContext(ctx)
}
```

### goroutinelint/testdata/src/goroutinelint/no_cancel.go

```go
package goroutinelint

func BadInfiniteGoroutine() {
    go func() { // want "AIL020: goroutine without cancellation" "AIL021: infinite loop"
        for {
            doWork()
        }
    }()
}

func GoodGoroutineWithContext(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                doWork()
            }
        }
    }()
}
```

### interfacelint/testdata/src/interfacelint/interfaces.go

```go
package interfacelint

// Bad: Too many methods - god interface
type Repository interface { // want "AIL040: interface has 7 methods"
    Create(item Item) error
    Read(id string) (Item, error)
    Update(item Item) error
    Delete(id string) error
    List() ([]Item, error)
    Search(query string) ([]Item, error)
    Count() (int, error)
}

// Good: Small, focused interfaces
type Reader interface {
    Read(id string) (Item, error)
}

type Writer interface {
    Write(item Item) error
}

// Bad: Single-method interface without "-er" suffix
type Validate interface { // want "AIL042: single-method interface"
    Validate() error
}

// Good: Proper naming
type Validator interface {
    Validate() error
}
```

### naminglint/testdata/src/naminglint/names.go

```go
package naminglint

type User struct {
    name string
}

// Bad: Go doesn't use Get prefix for getters
func (u *User) GetName() string { // want "AIL050: getter with Get prefix"
    return u.name
}

// Good: Just use Name()
func (u *User) Name() string {
    return u.name
}

// Bad: Redundant package prefix
type NaminglintService struct{} // want "AIL051: redundant package prefix"

// Good: Package provides context
type Service struct{}
```

### slicemaplint/testdata/src/slicemaplint/slices.go

```go
package slicemaplint

// Bad: Writing to nil map
func BadNilMap() {
    var m map[string]int
    m["key"] = 1 // want "AIL060: write to nil map"
}

// Good: Initialize map first
func GoodMap() {
    m := make(map[string]int)
    m["key"] = 1
}

// Bad: Modifying slice during iteration
func BadSliceModify(items []int) []int {
    for i, v := range items {
        if v < 0 {
            items = append(items[:i], items[i+1:]...) // want "AIL061: modifying slice"
        }
    }
    return items
}
```

### stringlint/testdata/src/stringlint/strings.go

```go
package stringlint

// Bad: Byte iteration on string
func BadByteIteration(s string) {
    for i := 0; i < len(s); i++ { // want "AIL070: byte iteration"
        _ = s[i]
    }
}

// Good: Rune iteration
func GoodRuneIteration(s string) {
    for _, r := range s {
        _ = r
    }
}

// Bad: String concatenation in loop
func BadConcat(items []string) string {
    var result string
    for _, s := range items {
        result += s // want "AIL071: string concat in loop"
    }
    return result
}

// Good: Use strings.Builder
func GoodConcat(items []string) string {
    var b strings.Builder
    for _, s := range items {
        b.WriteString(s)
    }
    return b.String()
}
```

### concurrencylint/testdata/src/concurrencylint/sync.go

```go
package concurrencylint

import "sync"

// Bad: wg.Done() not in defer
func BadWaitGroup(wg *sync.WaitGroup) {
    wg.Add(1)
    go func() {
        doWork()
        wg.Done() // want "AIL080: wg.Done not in defer"
    }()
}

// Good: Defer wg.Done()
func GoodWaitGroup(wg *sync.WaitGroup) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        doWork()
    }()
}

// Bad: Select with only default
func BadSelect(ch chan int) {
    select { // want "AIL082: select with only default"
    default:
        doWork()
    }
}
```

### paniclint/testdata/src/paniclint/panic.go

```go
package paniclint // This is a library package, not main

// Bad: Panic in library code
func ParseConfig(data []byte) Config {
    if len(data) == 0 {
        panic("empty config") // want "AIL090: panic in library"
    }
    return Config{}
}

// Good: Return error from library
func ParseConfigSafe(data []byte) (Config, error) {
    if len(data) == 0 {
        return Config{}, errors.New("empty config")
    }
    return Config{}, nil
}

// OK: Must pattern is acceptable
func MustParseConfig(data []byte) Config {
    cfg, err := ParseConfigSafe(data)
    if err != nil {
        panic(err) // OK - Must* functions can panic
    }
    return cfg
}
```

### initlint/testdata/src/initlint/init.go

```go
package initlint

import (
    "net/http"
    "os"
)

// Bad: Network call in init
func init() {
    resp, _ := http.Get("http://example.com") // want "AIL100: network in init"
    _ = resp
}

// Bad: File I/O in init
func init() {
    data, _ := os.ReadFile("config.json") // want "AIL101: file I/O in init"
    _ = data
}

// Good: Use explicit initialization
var config Config

func InitConfig() error {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return err
    }
    return json.Unmarshal(data, &config)
}
```

---

## Configuration

### .golangci.yml (for the project itself)

```yaml
version: "2"
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - errorlint
    - govet
    - staticcheck
    - unused
    - gocritic
    - revive
    - forcetypeassert
    - gosec

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
```

### go-ai-lint Configuration (for users)

Users will configure via `.go-ai-lint.yml` or `.golangci.yml`:

```yaml
# .go-ai-lint.yml
version: 1
analyzers:
  deferlint:
    enabled: true
    severity-override:
      AIL001: critical  # defer-in-loop
      AIL002: high      # defer-close-error-ignored
  contextlint:
    enabled: true
    allow-todo-in:
      - "**/*_test.go"
      - "**/test/**"
  goroutinelint:
    enabled: true
    min-go-version: "1.22"  # Skip loop var capture check if >= 1.22
  errorlint:
    enabled: true
  interfacelint:
    enabled: true
    max-methods: 5         # Default threshold for god interface
  naminglint:
    enabled: true
    skip-get-prefix: false # Set true to allow GetX() pattern
  slicemaplint:
    enabled: true
  stringlint:
    enabled: true
  concurrencylint:
    enabled: true
  paniclint:
    enabled: true
    allow-must-pattern: true  # Allow panic in MustX() functions
  initlint:
    enabled: true
    allow-file-io: false      # Strict mode
  optionlint:
    enabled: true
```

---

## Makefile

```makefile
.PHONY: all build test lint arch-check install clean

all: lint test build

build:
	go build -o bin/go-ai-lint ./cmd/go-ai-lint

test:
	go test -race -cover ./...

lint:
	golangci-lint run

arch-check:
	go-arch-lint check

install:
	go install ./cmd/go-ai-lint

clean:
	rm -rf bin/

# Development helpers
.PHONY: testdata coverage

testdata:
	@echo "Validating testdata..."
	go test ./internal/core/analyzers/... -v

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

---

## CI/CD (GitHub Actions)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: go build ./...

      - name: Test
        run: go test -race -cover ./...

      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

      - name: Architecture Check
        run: |
          go install github.com/fe3dback/go-arch-lint@latest
          go-arch-lint check
```

---

## Success Criteria

### v1.0.0 Release Criteria

- [ ] All 12 analyzer categories implemented
- [ ] At least 30 specific checks across categories
- [ ] Zero false positives on Go standard library
- [ ] < 5% false positive rate on real-world repos
- [ ] golangci-lint plugin working
- [ ] Documentation complete
- [ ] CI/CD pipeline green

### Analyzer Coverage Summary

| Analyzer | ID Range | Checks | Priority |
|----------|----------|--------|----------|
| deferlint | AIL001-009 | 5 | Sprint 1 |
| contextlint | AIL010-019 | 4 | Sprint 2 |
| goroutinelint | AIL020-029 | 4 | Sprint 3 |
| errorlint | AIL030-039 | 4 | Sprint 4 |
| interfacelint | AIL040-049 | 4 | Sprint 5 |
| naminglint | AIL050-059 | 4 | Sprint 5 |
| slicemaplint | AIL060-069 | 4 | Sprint 4 |
| stringlint | AIL070-079 | 3 | Sprint 6 |
| concurrencylint | AIL080-089 | 4 | Sprint 6 |
| paniclint | AIL090-099 | 3 | Sprint 7 |
| initlint | AIL100-109 | 3 | Sprint 7 |
| optionlint | AIL110-119 | 2 | Sprint 8 |
| **Total** | | **44** | |

### Quality Gates (Per PR)

- [ ] All tests pass
- [ ] Coverage >= 80%
- [ ] golangci-lint clean
- [ ] go-arch-lint clean
- [ ] No new false positives in testdata

---

## References

- [go/analysis Package](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- [golangci-lint Module Plugins](https://golangci-lint.run/docs/plugins/module-plugins/)
- [100 Go Mistakes](https://100go.co/)
- [NilAway Paper](https://www.uber.com/blog/nilaway-practical-nil-panic-detection-for-go/)
- [AST Explorer](https://astexplorer.net/)

---

## Notes for Implementation

When building this project with Claude:

1. **Always follow TDD** - Write failing tests first
2. **Check architecture** - Run `go-arch-lint check` after structural changes
3. **Run all quality gates** - `make all` before committing
4. **Use domain types** - Don't leak go/analysis types into core
5. **Keep analyzers focused** - One concern per analyzer
6. **Document edge cases** - In both code and testdata
7. **Minimize false positives** - Better to miss than to cry wolf
