# CLAUDE.md (Backend)

## Project Overview
AI Gateway is a Go-based backend that serves as an OpenAI-compatible proxy with multi-tenancy and telemetry support. It provides features for request routing, tenant context extraction, and performance monitoring.

## Build and Run Commands

### Commands (from /backend)
- Run development server: `make dev` (runs `go run ./cmd/gateway`)
- Build binary: `make build`
- Run tests: `make test`
- Run tests with coverage: `make test-cover`
- Lint: `make lint` (uses `golangci-lint`)
- Update dependencies: `make tidy`

### Docker Compose
- Start full stack: `make up`
- Tail logs: `make logs`
- Stop services: `make down`

## Coding Guidelines

### Go (Backend)
- **TDD Requirement**: Follow Test-Driven Development. Write failing tests before implementing features. Ensure high test coverage for the proxy logic.
- **Formatting**: Standard Go formatting (`gofmt`, `goimports`).
- **Logging**: Use `log/slog` with JSON handler as the default logger.
- **Routing**: Use `github.com/go-chi/chi/v5` for HTTP routing.
- **Telemetry**: Use OpenTelemetry (`go.opentelemetry.io`) for instrumentation.
- **Concurrency**: Use `golang.org/x/sync/errgroup` for managing groups of goroutines.
- **Error Handling**: Use `errors.Is` and `errors.As` for error checking. Wrap errors with context where appropriate.
- **Naming**: Use `PascalCase` for exported symbols and `camelCase` for internal variables/functions.
- **Architecture**: Keep the proxy "hot path" lean; use middleware for request enrichment (e.g., TenantContext).
