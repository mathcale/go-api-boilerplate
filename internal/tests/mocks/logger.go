package mocks

import (
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type Logger struct {
	mock.Mock
}

func (m *Logger) Info(msg string, kv ...any) {
	m.Called(msg, kv)
}

func (m *Logger) Warn(msg string, kv ...any) {
	m.Called(msg, kv)
}

func (m *Logger) Error(msg string, err error, kv ...any) {
	m.Called(msg, err, kv)
}

func (m *Logger) Debug(msg string, kv ...any) {
	m.Called(msg, kv)
}

func (m *Logger) Trace(msg string, kv ...any) {
	m.Called(msg, kv)
}

func (m *Logger) With(kv ...any) logger.Logger {
	args := m.Called(kv)
	return args.Get(0).(logger.Logger)
}

func (m *Logger) SetGlobalValue(key string, value any) {
	m.Called(key, value)
}

func (m *Logger) Instance() zerolog.Logger {
	args := m.Called()
	return args.Get(0).(zerolog.Logger)
}
