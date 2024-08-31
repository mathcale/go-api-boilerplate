# Go API Boilerplate

A slightly opinionated HTTP API boilerplate with the Go programming language, following (some) Clean Architecture principles.

[![Continuous Integration](https://github.com/mathcale/go-api-boilerplate/actions/workflows/ci.yaml/badge.svg)](https://github.com/mathcale/go-api-boilerplate/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mathcale/go-api-boilerplate)](https://goreportcard.com/report/github.com/mathcale/go-api-boilerplate)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.23-61CFDD.svg)

## Features

- HTTP server with [net/http](https://pkg.go.dev/net/http#hdr-Servers);
- Testing suites, assertions and mocks with [testify](https://github.com/stretchr/testify);
- Live reload with [air](https://github.com/air-verse/air);
- Logging with [zerolog](https://github.com/rs/zerolog);
- Configuration with [viper](https://github.com/spf13/viper);
- Pre-configured CI job with Github Actions;

## Requirements

- [Go](https://go.dev/) 1.23 (or newer)
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/)

## Running locally

```sh
# Start database
docker compose up -d

# Install dependencies
make setup

# Run the application
make run
```

## Testing

To execute all test suites, just run:

```sh
make tests
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

## Next Steps

- [X] Add database connection
- [X] Add logging middleware
- [ ] Add JWT user authentication
- [X] Add Github Actions CI workflow
