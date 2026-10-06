# Project Standards for Go Code

## Architecture Overview

This is a Go backend organised feature-first: a framework layer (`internal/core`) with no business
knowledge, plus vertical slices (`internal/modules/<name>`) that each follow Clean Architecture
internally. `cmd/api/main.go` is the composition root.

- **Core** (`internal/core/`): Framework and cross-cutting concerns shared by every module (config, logger, errors, JWT, bcrypt, database, email, middlewares, HTTP server, response helpers).
  - `core.Container` holds the shared collaborators (config, logger, DB, JWT, bcrypt, response, email) and is built once by `core.Build`.
  - `module.Module` is the contract every feature implements: `Name()` and `Routes() []module.Route`.
  - `core.CollectRoutes(mods...)` concatenates the routes of all modules; `httpserver.New` serves them.
- **Modules** (`internal/modules/<name>/`): Self-contained features (`user`, `system`, ...).
  - `domain/`: Business entities.
  - `gateway/`: The module's public port (`gateway.Gateway`), consumed by use cases and by other modules.
  - `usecases/`: Application business rules and orchestration; depends on `domain` and `gateway`.
  - `handlers/` (+ `handlers/dto/`): HTTP handlers and their request/response DTOs.
  - `mocks/`: Hand-written mocks of the module's public interfaces.
  - `internal/`: Module-private code, protected from other modules by the Go compiler: `models/` (DB models), `repository/` (raw SQL), `adapter/` (implements `gateway.Gateway` on top of the repository), `mappers/` (models <-> domain) and `mocks/` (mocks of private types).
  - `module.go` (`New(c *core.Container, ...)`) wires the module; `routes.go` declares its routes.

Key patterns:

- The gateway interface lives in `internal/modules/<name>/gateway` and is consumed directly by use cases.
- The adapter in `internal/modules/<name>/internal/adapter` implements the gateway and wraps its own repository (`repository.go`), raw SQL queries and `internal/models`.
- Each module wires its own repositories + adapter + use cases + handlers in `module.go`; `cmd/api/main.go` only builds the container, constructs the modules and passes them to `core.CollectRoutes`.
- Structured error handling with the `apperror` package using business codes; errors always preserve their underlying cause (implement `Unwrap`).
- Structured logging with zerolog and context-aware log fields.

## File Structure

```text
cmd/api/main.go            Composition root
internal/core/             Framework (config, logger, apperror, apierror, jwt, bcrypt, database,
                           email, webcontext, middlewares, response, module, httpserver, container.go)
internal/modules/<name>/   Feature slices (domain, gateway, usecases, handlers, mocks, internal/...)
internal/tests/mocks/      Shared mocks of core collaborators (logger, jwt, password, email client)
migrations/                SQL migrations
docs/                      Generated Swagger output
```

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
  - The gateway interface belongs to the module in `internal/modules/<name>/gateway` (type `Gateway`).
  - The implementation (`internal/adapter`) and its repository (`internal/repository`) are private to the module.
- **Dependency Injection**: Wire dependencies by hand in each module's `module.go`, and compose modules in `cmd/api/main.go`.
- **Security**: Validate inputs at the boundary and always use parameterized SQL.
- **SQL queries**: Use sqlx safe parameter binding. Keep raw SQL as constants alongside the repository in `internal/repository/repository.go`.
- **Comments**: Do NOT add comments to code, unless it is an HTTP handler (in that case, add OpenAPI/Swagger annotations based on `godoc`) or a non-obvious decision worth recording.

## Project-Specific Patterns

### Use Case Pattern

- Use cases depend on the module's gateway interface.
- Example: `internal/modules/user/usecases/signup.go` uses `internal/modules/user/gateway`.

### Gateway Pattern

- The gateway interface lives in `internal/modules/<name>/gateway/gateway.go`.
- The implementation lives in `internal/modules/<name>/internal/adapter` and wraps the module's repository.
- Example: `internal/modules/user/internal/adapter/adapter.go`.

### Module Wiring

- Each module's `New(c *core.Container, ...)` in `module.go` builds its repository, adapter, use cases and handlers.
- Routes are declared in `routes.go` as `[]module.Route` (use `module.Public/Protected` and `module.RateLimited/NotThrottled`).
- Modules are composed in `cmd/api/main.go`.

### Middleware Chain

- Global middlewares are ordered by the resolver in `internal/core/httpserver/middlewares.go`.
- Per-route protection (authentication, RBAC, rate limiting) is applied in `internal/core/httpserver/server.go` based on the flags of each `module.Route`.

### Authentication & RBAC

- Access and refresh tokens are separate, signed with distinct secrets (`internal/core/jwt`).
- The authenticated user's ID and roles are carried through the request via `internal/core/webcontext`.
- Routes requiring a role set `RequiredRoles` on the `module.Route`; the `RequireRole` middleware enforces them after authentication. See the `GET /v1/admin/ping` demo route in `internal/modules/system/routes.go`.

## Adding a module

1. Create `internal/modules/<name>/` with `domain/`, `gateway/`, `usecases/`, `handlers/` (+ `dto/`), `mocks/` and `internal/{models,repository,adapter,mappers,mocks}`.
2. Add `module.go` with `New(c *core.Container, deps...)` returning a type that implements `module.Module`, and `routes.go` with its `Routes()`.
3. In `cmd/api/main.go`, call `<name>.New(c)` and add the module to `core.CollectRoutes(...)`.
4. Name new migrations `NNNNNN_<name>_<desc>`.

### Import rules

- A module imports only `internal/core/*` and the non-`internal` packages (`domain`, `gateway`) of other modules.
- Cross-module access goes only through another module's `gateway.Gateway` port, passed into `New`.
- `internal/core` never imports `internal/modules` (enforced by `depguard`).
- Modules do not share DTOs.

## Database schema

The database structure evolves through migrations located in the `./migrations` folder. To understand the current schema, read all migration files there.
