package mocks

import (
	"github.com/stretchr/testify/mock"
)

type LoggerMock struct {
	mock.Mock
}

func (m *LoggerMock) Info(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *LoggerMock) Warn(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *LoggerMock) Error(msg string, err error, tags map[string]interface{}) {
	m.Called(msg, err, tags)
}

func (m *LoggerMock) Debug(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}

func (m *LoggerMock) Trace(msg string, tags map[string]interface{}) {
	m.Called(msg, tags)
}
