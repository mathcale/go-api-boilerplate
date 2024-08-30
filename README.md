# Go API Boilerplate

A slightly opinionated HTTP API boilerplate with the Go programming language, following (some) Clean Architecture principles.

![Go Version](https://img.shields.io/badge/go%20version-%3E=1.23-61CFDD.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/mathcale/go-api-boilerplate)](https://goreportcard.com/report/github.com/mathcale/go-api-boilerplate)

## Features

- HTTP server with [net/http](https://pkg.go.dev/net/http#hdr-Servers);
- Testing suites, assertions and mocks with [testify](https://github.com/stretchr/testify);
- Live reload with [air](https://github.com/air-verse/air);
- Logging with [zerolog](https://github.com/rs/zerolog);
- Configuration with [viper](https://github.com/spf13/viper);

## Running

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

## Next Steps

- [X] Add database connection
- [ ] Add JWT user authentication
- [-] Add logging and tracing middlewares
