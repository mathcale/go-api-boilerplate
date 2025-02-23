package mocks

import (
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

type Logger struct {
	mock.Mock
}

func (m *Logger) Info(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *Logger) Warn(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *Logger) Error(msg string, err error, tags map[string]interface{}) {
	m.Called(msg, err, tags)
}

func (m *Logger) Debug(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *Logger) Trace(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *Logger) SetGlobalValue(key string, value any) {
	m.Called(key, value)
}

func (m *Logger) Instance() zerolog.Logger {
	args := m.Called()
	return args.Get(0).(zerolog.Logger)
}
