<!-- markdownlint-disable MD029 -->
# Go API Boilerplate

A slightly opinionated HTTP API boilerplate with the Go programming language, following (some) Clean Architecture principles.

[![Continuous Integration](https://github.com/mathcale/go-api-boilerplate/actions/workflows/ci.yaml/badge.svg)](https://github.com/mathcale/go-api-boilerplate/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mathcale/go-api-boilerplate)](https://goreportcard.com/report/github.com/mathcale/go-api-boilerplate)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.24-61CFDD.svg)
[![MIT License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)

## Features

- HTTP server with [net/http](https://pkg.go.dev/net/http#hdr-Servers);
- Testing suites, assertions and mocks with [testify](https://github.com/stretchr/testify);
- Live reload with [air](https://github.com/air-verse/air);
- Logging with [zerolog](https://github.com/rs/zerolog);
- Configuration with [viper](https://github.com/spf13/viper);
- PostgreSQL database connection with [pgx](https://github.com/jackc/pgx) and [sqlx](https://github.com/jmoiron/sqlx)
- Pre-configured CI job with Github Actions;
- JWT authentication (under the [`with-auth`](../../tree/with-auth) branch)

## Project structure

```plaintext
.
├── cmd
│   └── api                    <== Application entrypoint
├── config                     <== Configurations setup
├── docs                       <== Auto-generated OpenAPI documents
├── internal
│   ├── domain                 <== Domain objects with their own contracts and business logic
│   ├── infra                  <== Everything related to the infrastructure layer, such as databases, web handlers, queues etc
│   │   ├── database
│   │   │   ├── models         <== Objects representing database entities
│   │   │   └── repositories   <== Database operations and queries
│   │   ├── gateways           <== Objects that encapsulates access to infra resources, following an interface defined by each use-case
│   │   └── web                <== Everything related to the HTTP REST API
│   │       ├── handlers       <== "Controllers"
│   │       │   └── dto        <== Input and output objects
│   │       └── middlewares    <== Request middlewares
│   ├── pkg                    <== Shared code that doesn't contain business logic, but are needed to support other packages
│   │   ├── apierror           <== Error returned to the API client
│   │   ├── apperror           <== Internal error object
│   │   ├── di                 <== Dependency injection resolver, where everything is glued together
│   │   └── logger             <== Logging utilities
│   ├── tests                  <== Testing utilities
│   │   ├── fixtures           <== Pre-built objects for test cases
│   │   └── mocks              <== Methods mocks for necessary packages
│   └── usecases               <== Main business rules
│       └── counter
├── migrations                 <== Database migrations generated with `migrate`
├── scripts                    <== Utilitarian shell scripts
```

## Requirements

- [Go](https://go.dev/) 1.24 (or newer)
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/)
- [air](https://github.com/air-verse/air): live-reloading
- [migrate](https://github.com/golang-migrate/migrate): database migrations

## First run

1. Create .env file

```sh
cp .env.example .env
```

2. Rename packages to your project's name

```sh
make rename-pkgs
```

3. Run setup script

```sh
make setup
```

## Running locally

To start a local server with live-reload and all necessary containers, execute:

```sh
make run
```

## Testing

To execute all test suites and get a coverage report at the end, just run:

```sh
make test
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

By running the following command, the application will be compiled and outputted to the `bin` directory.

```sh
make build
```

## Next Steps

- [X] Add database connection
- [X] Add logging middleware
- [X] Add Github Actions CI workflow
- [X] Add database
- [X] Add authentication branch
- [X] Add OpenAPI specs on web handlers
- [X] Improve project structure documentation
