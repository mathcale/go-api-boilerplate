package mocks

import (
	"github.com/rs/zerolog"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

// NoopLogger is a logger.Logger that discards everything. Use it in use-case and
// handler tests where log output is irrelevant, to avoid setting up expectations
// on every logging call.
type NoopLogger struct{}

func (NoopLogger) Info(string, ...any)         {}
func (NoopLogger) Warn(string, ...any)         {}
func (NoopLogger) Error(string, error, ...any) {}
func (NoopLogger) Debug(string, ...any)        {}
func (NoopLogger) Trace(string, ...any)        {}
func (n NoopLogger) With(...any) logger.Logger { return n }
func (NoopLogger) SetGlobalValue(string, any)  {}
func (NoopLogger) Instance() zerolog.Logger    { return zerolog.Nop() }
