<!-- markdownlint-disable MD029 -->
# Go API Boilerplate

A slightly opinionated HTTP API boilerplate with the Go programming language, following (some) Clean Architecture principles.

[![Continuous Integration](https://github.com/mathcale/go-api-boilerplate/actions/workflows/ci.yaml/badge.svg)](https://github.com/mathcale/go-api-boilerplate/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mathcale/go-api-boilerplate)](https://goreportcard.com/report/github.com/mathcale/go-api-boilerplate)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.27-61CFDD.svg)

## Features

- HTTP server with [net/http](https://pkg.go.dev/net/http#hdr-Servers);
- Testing suites, assertions and mocks with [testify](https://github.com/stretchr/testify);
- Live reload with [air](https://github.com/air-verse/air);
- Logging with [zerolog](https://github.com/rs/zerolog);
- Configuration with [viper](https://github.com/spf13/viper);
- PostgreSQL database connection with [pgx](https://github.com/jackc/pgx) and [sqlx](https://github.com/jmoiron/sqlx)
- Pre-configured CI job with Github Actions;

## Requirements

- [Go](https://go.dev/) 1.27 (or newer)
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/)
- [air](https://github.com/air-verse/air): live-reloading
- [migrate](https://github.com/golang-migrate/migrate): database migrations
- [golangci-lint](https://golangci-lint.run/): linting
- [swag](https://github.com/swaggo/swag): OpenAPI generation

## Running locally

1. Create .env file

```sh
cp .env.example .env
```

Then fill in `ACCESS_TOKEN_SECRET` and `REFRESH_TOKEN_SECRET` with distinct values of at
least 64 bytes (e.g. `openssl rand -base64 64`). The API refuses to start otherwise.

Access tokens are short-lived (`ACCESS_TOKEN_LIFETIME_MINUTES`, 15 by default). Each user carries a
`token_version` that is bumped on password change; tokens issued with an older version are rejected.

Email delivery is controlled by `MAILER_PROVIDER`:

- `fake` (default, local development only): emails are not sent. Rejected when
  `ENVIRONMENT=production`.
- `smtp`: sends through an SMTP server over an encrypted connection. Requires `SMTP_HOST` and
  `SMTP_PORT`; `SMTP_USERNAME`/`SMTP_PASSWORD` enable PLAIN authentication, and `SMTP_TLS_MODE` is
  `starttls` (default, e.g. port 587) or `implicit` (e.g. port 465).

Rate limiting is keyed by client IP (IPv6 clients by their /64). By default only the TCP peer
address is used. When running behind a reverse proxy or load balancer, set `TRUSTED_PROXIES` to a
comma-separated list of the proxies' IPs/CIDRs (e.g. `10.0.0.0/8,192.168.1.5`); `X-Forwarded-For`
is then honoured only for requests coming from those addresses, taking the first untrusted entry
from the right. Leave it empty otherwise, or clients could spoof their address.

Sign-in attempts are additionally limited per account email (`SIGNIN_RATE_LIMIT_PER_MINUTE` and
`SIGNIN_RATE_LIMIT_BURST`, 5 and 5 by default).

> Rate limits are kept in process memory and are not shared across replicas: with several
> instances each one enforces its own limits. Use a shared store such as Redis if you need limits
> that hold across replicas.

2. Rename packages to your project's name

```sh
make rename-pkgs
```

3. Run setup script

```sh
make setup
```

4. Start server

```sh
make run
```

Interactive API docs are then available at `http://localhost:8000/swagger/`.

## Testing

To execute all tests with a coverage report, run:

```sh
make test
```

## Linting

```sh
make lint      # report issues
make lint-fix  # auto-fix where possible
make fmt       # format imports and code
```

## Building for production

### With Docker

There's a `Dockerfile.prod` included with the project to build an optimized image based on [distroless](https://github.com/GoogleContainerTools/distroless), so you just need to adapt it for your needs and publish to your desired registry.

```sh
# This should be set by your CI/CD system
export BUILD_ID="$(uuidgen)"

# Building the image
docker build . \
  -t mathcale/go-api-boilerplate \
  -f Dockerfile.prod \
  --build-arg BUILD_ID

# Clean intermediate images
docker image prune \
  --filter label=stage=builder \
  --filter label=build=$BUILD_ID
```

### Manually

```sh
make build
```
