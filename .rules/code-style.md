# Project Standards for Go Code

## Architecture Overview

This is a Go backend following Clean Architecture principles with clear separation of concerns:

- **Domain** (`internal/domain/`): Core business entities and models, grouped by bounded context (e.g. `user`, ...)
- **Domain Gateways** (`internal/domain/<context>/gateway/`): Gateway *interfaces* (ports) owned by the domain/bounded context
- **Use Cases** (`internal/usecases/`): Application business rules and orchestration; depends on domain entities and domain gateway interfaces
- **Infrastructure** (`internal/infra/`): External concerns (web, database, email)
  - `internal/infra/gateways/<context>/`: Gateway *implementations* (adapters), including their repositories and raw SQL queries
  - `internal/infra/database/`: SQL/DB connection and models
  - `internal/infra/web/`: HTTP server, routes, handlers and middlewares
  - `internal/infra/email/`: Email clients and provider integration
- **Shared Packages** (`internal/pkg/`): Reusable utilities and cross-cutting concerns

Key patterns:

- Domain gateway interfaces live in `internal/domain/<context>/gateway` and are consumed directly by use cases.
- Gateway implementations live in `internal/infra/gateways/<context>` and include their own repository (`repository.go`), wrapping raw SQL queries and `internal/infra/database/models`.
- Dependency injection wires repositories + gateway implementations + use cases in `internal/pkg/di/`.
- Structured error handling with the `apperror` package using business codes; errors always preserve their underlying cause (implement `Unwrap`).
- Structured logging with zerolog and context-aware log fields.

## File Structure

All core Go files should be organized inside the `internal` directory, in a way that reflects the project's architecture.

- `domain`: Core business models per bounded context.
  - `<context>/gateway`: Domain gateway interfaces (ports).
- `usecases`: Orchestration and business rules. Use cases depend on domain models and domain gateway interfaces.
- `infra`: External concerns.
  - `database`: Connection and models.
  - `gateways`: Gateway implementations (adapters) that implement domain gateway interfaces, along with their repositories.
  - `web`: HTTP server setup, routes, handlers and middlewares.
  - `email`: Email clients and provider integration.
- `pkg`: Cross-cutting shared utilities.
  - `apierror`: Errors for API responses.
  - `apperror`: Application errors used by internal layers.
  - `bcrypt`: Password hashing utilities.
  - `di`: Dependency injection.
  - `jwt`: JWT utilities.
  - `logger`: Logging.
  - `mappers`: Conversions between infra/database models and domain objects.
- `tests`: Test utilities, mocks and fixtures.

## Critical Developer Workflows

### Development Setup

```bash
make setup  # Installs deps, starts containers, initializes DB, runs migrations
```

### Local Development

```bash
make run    # Starts live-reload server with air + containers
```

### Testing

```bash
make test   # Runs all tests with coverage report (HTML output)
```

### Database Operations

```bash
make create-migration  # Creates new migration file
make migrate-up        # Applies pending migrations
make migrate-down      # Rolls back migrations
```

### Code Quality

```bash
make lint       # Run golangci-lint
make lint-fix   # Auto-fix linting issues
make fmt        # Format code
```

### Building

```bash
make build      # Production build to ./bin/api
```

## Implementation Guidelines

- **Error Handling**: Use Go's error handling idioms. Prefer returning errors over panicking, except in truly exceptional situations. Always wrap the underlying error via `apperror.New(err, ...)` so the cause chain is preserved.
- **Logging**: Use structured logging for better log management and analysis. Ensure that sensitive information (passwords, tokens, codes) is never logged.
- **Testing**: Write unit tests for functions and methods using `testify` with hand-written mocks. See `.rules/testing.md`.
- **Code style**: Follow Go conventions with `gofmt` / `golangci-lint`. Keep lines within 100 columns. Use tabs for indentation.
- **Use-case Driven Development**: Structure your code around use cases.
- **Gateways and Repositories**:
  - Gateway interfaces belong to the domain bounded context in `internal/domain/<context>/gateway`.
  - Gateway implementations and their repositories belong to infrastructure in `internal/infra/gateways/<context>`.
- **Dependency Injection**: Wire dependencies in `internal/pkg/di/`.
- **Security**: Validate inputs at the boundary and always use parameterized SQL.
- **SQL queries**: Use sqlx safe parameter binding. Keep raw SQL as constants alongside the repository in `repository.go`.
- **Comments**: Do NOT add comments to code, unless it is an HTTP handler (in that case, add OpenAPI/Swagger annotations based on `godoc`) or a non-obvious decision worth recording.

## Project-Specific Patterns

### Use Case Pattern

- Use cases depend on domain gateway interfaces.
- Example: `internal/usecases/auth/signup.go` uses `internal/domain/user/gateway`.

### Domain Gateway Pattern

- Domain gateway interfaces live in `internal/domain/<context>/gateway`.
- Example: `internal/domain/user/gateway/interface.go`.

### Infra Gateway Pattern

- Infra gateway implementations live in `internal/infra/gateways/<context>`.
- They implement the domain gateway interfaces and wrap repositories.
- Example: `internal/infra/gateways/user/gateway.go`.

### Dependency Injection

- DI wires repositories and gateway implementations in `internal/pkg/di/inject.go`.

### Middleware Chain

- Global middlewares are ordered by the resolver in `internal/infra/web/middlewares.go`.
- Per-route protection (authentication, RBAC, rate limiting) is applied in `internal/infra/web/server.go` based on flags in `internal/infra/web/router.go`.

### Authentication & RBAC

- Access and refresh tokens are separate, signed with distinct secrets (`internal/pkg/jwt`).
- The authenticated user's ID and roles are carried through the request via `internal/infra/web/webcontext`.
- Routes requiring a role set `requiredRoles` in the router; the `RequireRole` middleware enforces them after authentication. See the `GET /v1/admin/ping` demo route.

## Database schema

The database structure evolves through migrations located in the `./migrations` folder. To understand the current schema, read all migration files there.
