package mocks

import "github.com/rs/zerolog"

// NoopLogger is a logger.Logger that discards everything. Use it in use-case and
// handler tests where log output is irrelevant, to avoid setting up expectations
// on every logging call.
type NoopLogger struct{}

func (NoopLogger) Info(string, map[string]interface{})         {}
func (NoopLogger) Warn(string, map[string]interface{})         {}
func (NoopLogger) Error(string, error, map[string]interface{}) {}
func (NoopLogger) Debug(string, map[string]interface{})        {}
func (NoopLogger) Trace(string, map[string]interface{})        {}
func (NoopLogger) SetGlobalValue(string, any)                  {}
func (NoopLogger) Instance() zerolog.Logger                    { return zerolog.Nop() }
