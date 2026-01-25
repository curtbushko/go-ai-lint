# go-ai-lint

A static analysis tool for detecting common mistakes in AI-generated Go code.

## Overview

`go-ai-lint` catches issues that AI code generators (ChatGPT, Copilot, Claude) frequently produce incorrectly. Each diagnostic includes rich guidance (Why/Fix/Example) designed to help AI assistants avoid fix loops.

## Installation

```bash
go install github.com/curtbushko/go-ai-lint/cmd/go-ai-lint@latest
```

## Usage

### Standalone CLI

```bash
# Run all analyzers
go-ai-lint ./...

# Run specific analyzers
go-ai-lint -deferlint -errorlint ./...

# Show help
go-ai-lint -help
```

### With golangci-lint (Plugin)

Add to your `.golangci.yml`:

```yaml
linters-settings:
  custom:
    go-ai-lint:
      path: github.com/curtbushko/go-ai-lint
      description: Detects common AI-generated Go code mistakes
      original-url: github.com/curtbushko/go-ai-lint
```

## Analyzers

| Analyzer | ID Range | Description |
|----------|----------|-------------|
| deferlint | AIL001-009 | Defer mistakes (in-loop, ignored errors, nil receiver) |
| contextlint | AIL010-019 | Context misuse (TODO in production, Background in handler) |
| goroutinelint | AIL020-029 | Goroutine lifecycle (no cancel, infinite loop, closure capture) |
| errorlint | AIL030-039 | Error handling (handled twice, %v instead of %w) |
| interfacelint | AIL040-049 | Interface design (too large, missing -er suffix) |
| naminglint | AIL050-059 | Naming conventions (Get prefix, redundant package name) |
| slicemaplint | AIL060-069 | Slice/map pitfalls (nil map write, modify during iteration) |
| stringlint | AIL070-079 | String handling (byte iteration, concat in loop) |
| concurrencylint | AIL080-089 | Concurrency issues (WaitGroup.Done not deferred, select only default) |
| paniclint | AIL090-099 | Panic/recover (panic in library, empty recover) |
| initlint | AIL100-109 | Init function issues (network calls, file I/O) |
| optionlint | AIL110-119 | Functional options pattern (With* not returning Option) |

## Diagnostic Examples

### AIL001: defer-in-loop

```go
// Bad: defer accumulates until function returns
for _, f := range files {
    file, _ := os.Open(f)
    defer file.Close() // AIL001: defer inside loop
}

// Good: extract to helper function
for _, f := range files {
    processFile(f)
}

func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    return process(file)
}
```

### AIL030: error-handled-twice

```go
// Bad: error logged AND returned
if err != nil {
    log.Printf("error: %v", err)
    return err // AIL030: error handled twice
}

// Good: wrap and return only
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### AIL060: nil-map-write

```go
// Bad: writing to nil map panics
var m map[string]int
m["key"] = 1 // AIL060: write to nil map

// Good: initialize first
m := make(map[string]int)
m["key"] = 1
```

### AIL080: waitgroup-done-not-deferred

```go
// Bad: Done() may not run if goroutine panics
go func() {
    doWork()
    wg.Done() // AIL080: wg.Done() should be deferred
}()

// Good: defer guarantees Done() runs
go func() {
    defer wg.Done()
    doWork()
}()
```

## AI-Friendly Output

Use the AI reporter for rich diagnostics that prevent fix loops:

```go
import "github.com/curtbushko/go-ai-lint/internal/adapters/reporters"

reporter := reporters.NewAIReporter(os.Stdout)
reporter.Report(issues)
```

Output includes:
- **why**: Consequence of the problem
- **fix**: Strategy to resolve it
- **example**: Before/after code
- **common_mistakes**: What NOT to do when fixing

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Build CLI
make build

# Run all checks
make all
```

## Architecture

Hexagonal architecture with clean separation:

```
internal/
├── core/
│   ├── domain/      # Domain types (Issue, Severity, Category)
│   ├── ports/       # Interfaces (Analyzer, Reporter)
│   └── analyzers/   # Individual analyzer implementations
└── adapters/
    └── reporters/   # Output format adapters (Text, JSON, AI)
```

## License

MIT
