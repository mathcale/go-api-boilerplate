# Testing

## Strategy

- Write unit tests for use cases, packages (`jwt`, `bcrypt`), HTTP handlers, and repositories. Aim to cover the happy path plus the meaningful failure branches (validation, not-found, conflict, unauthorized, forbidden, expired).
- Tests live next to the code they exercise, in an external `_test` package (e.g. `package auth_test`) so they consume the public API only.

## Tooling

- Use `testify` (`assert` / `require`) for assertions. Prefer `require` for fatal preconditions and `assert` for follow-up checks.
- Use hand-written `testify/mock` doubles kept in `internal/tests/mocks/` rather than a code-generation tool, so mocks stay readable and version-controlled. There is a mock for each collaborator interface (gateway, password, jwt, logger, use cases).
- `mocks.NoopLogger{}` is a no-op `logger.Logger` for tests that don't assert on logging.
- Repository tests use `github.com/DATA-DOG/go-sqlmock` wrapped in an `sqlx.DB` to assert on the exact SQL, arguments, and transaction boundaries.
- Handler tests use `net/http/httptest` to drive the handler directly and assert on status codes and JSON bodies.

## Conventions

- Name tests `Test<Subject>_<Scenario>` (e.g. `TestSignInUseCase_WrongPassword`).
- Assert on the `apperror.BusinessCode` and HTTP status for error branches so a regression in the error contract is caught.
- Keep mocks strict: set expectations with `mock.On(...)` and verify with `AssertExpectations` / `AssertNotCalled` to prove collaborators were (not) invoked.

## Commands

```bash
make test   # Runs all tests with a coverage report
go test ./internal/...  # Quick run of the internal packages
```
