# AGENTS.md - AI Coding Agent Instructions

This document provides guidelines for AI coding agents working in the `github.com/alexhokl/helper` repository.

## Project Overview

A Go utility library (Go 1.25.5) providing reusable helper packages for:
- Database operations (PostgreSQL, SQL Server)
- API integrations (Airtable, Google APIs, Strava, GitHub)
- Authentication/OAuth helpers
- CLI helpers (Cobra/Viper)
- Telemetry/observability (OpenTelemetry)
- General utilities (collections, datetime, regex, crypto, JSON, IO)

## Build/Lint/Test Commands

This project uses [Task](https://taskfile.dev/) as the build runner. Commands:

| Command | Description |
|---------|-------------|
| `task build` | Go build without output (`go build ./...`) |
| `task test` | Run all unit tests (`go test ./...`) |
| `task coverage` | Unit tests with coverage (`go test --cover ./...`) |
| `task coverage-html` | Generate coverage HTML report |
| `task lint` | Run linter (`golangci-lint run`) |
| `task sec` | Security check (`gosec ./...`) |
| `task bench` | Run benchmarks (`go test -bench=. -benchmem ./...`) |

### Running a Single Test

```bash
# Run a specific test function
go test -v -run TestFunctionName ./package/

# Examples:
go test -v -run TestGetDistinct ./collection/
go test -v -run TestGetCurrentBranchName ./git/
go test -v -run TestIsFileExist ./iohelper/

# Run all tests in a specific package
go test -v ./collection/

# Run a specific benchmark
go test -bench=BenchmarkGetDelimitedString -benchmem ./collection/

# Run tests with race detection
go test -race -v -run TestFunctionName ./package/
```

## Code Style Guidelines

### Import Organization

Use three groups separated by blank lines:
1. Standard library
2. External third-party packages
3. Internal packages (when applicable)

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)
```

### Naming Conventions

| Element | Convention | Examples |
|---------|------------|----------|
| Files | lowercase_with_underscores | `open_telemetry.go`, `log_handler.go` |
| Packages | short lowercase; `*helper` suffix for utilities | `cli`, `httphelper`, `jsonhelper` |
| Exported functions | PascalCase, verb-based | `GetCurrentBranchName`, `WriteToJSONFile` |
| Unexported functions | camelCase | `execute`, `getPemBlock` |
| Variables | camelCase | `branchName`, `connectionString` |
| Constants | PascalCase (exported), camelCase (unexported) | `TraceIDKey`, `apiURL` |
| Structs/Types | PascalCase nouns | `Config`, `PostgresConfig`, `GithubIssue` |

### Error Handling

1. Always check errors immediately after function calls
2. Wrap errors with context using `fmt.Errorf` and `%w`:
```go
return nil, fmt.Errorf("unable to read file %s: %w", path, err)
```

3. Validate inputs early and return descriptive errors:
```go
if path == "" {
    return nil, fmt.Errorf("path cannot be empty")
}
```

4. Use `defer` for resource cleanup:
```go
file, err := os.Open(path)
if err != nil {
    return nil, err
}
defer file.Close()
```

### Context Usage

Pass `context.Context` as the first parameter for async operations:
```go
func GetIssue(ctx context.Context, client *githubv4.Client, owner, repo string) (*GithubIssue, error)
```

### Logging

Use `log/slog` for structured logging:
```go
slog.Error("unable to process request",
    slog.String("path", path),
    slog.String("error", err.Error()),
)
```

### Documentation

Use GoDoc-style comments starting with the function name:
```go
// GetOpenCommand returns the command and its required arguments according to
// the current operating system.
func GetOpenCommand(args ...string) (string, []string) {
```

### Testing

- **Table-driven tests** with `t.Run()` for subtests
- **Helper functions** should call `t.Helper()`
- **Benchmarks** use `func Benchmark*(b *testing.B)`
- Use `t.TempDir()` for temporary files and `t.Setenv()` for environment variables

### Struct Tags

Use tag ordering: `yaml`, `json`, `mapstructure`, `env`:
```go
Server string `yaml:"server" json:"server" env:"DB_SERVER"`
```

## Key Dependencies

- **CLI**: `github.com/spf13/cobra`, `github.com/spf13/viper`
- **Database**: `gorm.io/gorm`, `gorm.io/driver/postgres`
- **TUI**: `github.com/charmbracelet/bubbletea`
- **OAuth**: `golang.org/x/oauth2`
- **Google APIs**: `google.golang.org/api`
- **OpenTelemetry**: `go.opentelemetry.io/otel/*`
- **GitHub**: `github.com/shurcooL/githubv4`
- **LLM**: `github.com/tmc/langchaingo`

## Before Committing

1. Run `task lint` - fix any linter errors
2. Run `task test` - ensure all tests pass
3. Run `task sec` - check for security issues
4. Add tests for new functionality
5. Update GoDoc comments for exported functions