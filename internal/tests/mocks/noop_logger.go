package mocks

import (
	"github.com/rs/zerolog"

	"go-api-boilerplate/internal/pkg/logger"
)

type NoopLogger struct{}

func (NoopLogger) Info(string, ...any) {}

func (NoopLogger) Warn(string, ...any) {}

func (NoopLogger) Error(string, error, ...any) {}

func (NoopLogger) Debug(string, ...any) {}

func (NoopLogger) Trace(string, ...any) {}

func (n NoopLogger) With(...any) logger.Logger {
	return n
}

func (NoopLogger) Instance() zerolog.Logger {
	return zerolog.Nop()
}
