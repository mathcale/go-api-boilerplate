# Testing

## Strategy

- Write unit tests for use cases, packages (`jwt`, `bcrypt`), HTTP handlers, and repositories. Aim to cover the happy path plus the meaningful failure branches (validation, not-found, conflict, unauthorized, forbidden, expired).
- Tests live next to the code they exercise, in the **same package** as that code (e.g. `package user`), never in an external `_test` package (`package user_test`).
- **Exception (import cycles):** `internal/tests/mocks` and `internal/tests/fixtures` import domain, `jwt` and use-case packages. A test in the same package as one of those would form an import cycle, so tests that use mocks/fixtures for such a package must stay in the external `_test` package. Currently: `internal/domain/user`, `internal/pkg/jwt` (except unexported-only tests, which use `package jwt` without fixtures) and `internal/usecases/auth`. Keep the exception list in sync when dependencies change; never add a new `_test` package otherwise.

## Tooling

- Use `testify` suites (`testify/suite`). Assert through the suite (`s.Equal`, `s.True`, ...); use `s.Require()` for fatal preconditions and the plain assertions for follow-up checks.
- Use hand-written `testify/mock` doubles rather than a code-generation tool, so mocks stay readable and version-controlled. There is a mock for each collaborator interface (gateway, password, jwt, logger, use cases).
- **All mocks live in `internal/tests/mocks/`**, never next to the code they mock.
- **All object and network-response fixtures live in `internal/tests/fixtures/`.** Tests build their inputs and expected payloads from there instead of declaring them inline.
- `mocks.NoopLogger{}` is a no-op `logger.Logger` for tests that don't assert on logging.
- Repository tests use `github.com/DATA-DOG/go-sqlmock` wrapped in an `sqlx.DB` to assert on the exact SQL, arguments, and transaction boundaries.
- Handler tests use `net/http/httptest` to drive the handler directly and assert on status codes and JSON bodies.

## Suite structure

Reference example: `internal/domain/user/refreshtoken_test.go`.

- Each test file declares one suite struct embedding `suite.Suite`, named `<Subject>TestSuite` (e.g. `RefreshTokenDomainTestSuite`).
- Each suite has a single **main runner**, a plain `Test<Subject>` function that does nothing but call `suite.Run(t, new(<Subject>TestSuite))`.
- Each method of the subject under test gets **its own suite method** named `Test<Method>` (e.g. `TestRotateRefreshToken`). Never cover two methods in one suite method.
- Inside a suite method, every scenario is a subtest in its own `s.Run("should ...", func() { ... })`.
- Order the subtests **success paths first, then failure scenarios**.
- Use `SetupTest` to build fresh mocks per test when the suite has collaborators.

```go
type RefreshTokenDomainTestSuite struct {
	suite.Suite
}

func TestRefreshTokenDomain(t *testing.T) {
	suite.Run(t, new(RefreshTokenDomainTestSuite))
}

func (s *RefreshTokenDomainTestSuite) TestIsExpired() {
	s.Run("should report an unexpired token as not expired", func() { /* ... */ })
	s.Run("should report a past token as expired", func() { /* ... */ })
}
```

## Conventions

- Assert on the `apperror.BusinessCode` and HTTP status for error branches so a regression in the error contract is caught.
- Keep mocks strict: set expectations with `mock.On(...)` and verify with `AssertExpectations` / `AssertNotCalled` to prove collaborators were (not) invoked.

## Commands

```bash
make test   # Runs all tests with a coverage report
go test ./internal/...  # Quick run of the internal packages
```
