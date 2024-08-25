# Go API Boilerplate

A slightly opinionated HTTP API boilerplate with the Go programming language, following (some) Clean Architecture principles.

## Features

- HTTP server with [net/http](https://pkg.go.dev/net/http#hdr-Servers);
- Testing suites, assertions and mocks with [testify](https://github.com/stretchr/testify);
- Live reload with [air](https://github.com/air-verse/air);
- Logging with [zerolog](https://github.com/rs/zerolog);
- Configuration with [viper](https://github.com/spf13/viper);

## Running

```bash
# Install dependencies
make setup

# Run the application
make run
```

## Next Steps

- [X] Add database connection
- [ ] Add JWT user authentication
- [ ] Add logging and tracing middlewares
