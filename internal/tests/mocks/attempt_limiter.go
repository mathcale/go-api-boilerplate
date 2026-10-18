package mocks

import "github.com/stretchr/testify/mock"

type AttemptLimiter struct {
	mock.Mock
}

func (m *AttemptLimiter) Allow(key string) bool {
	return m.Called(key).Bool(0)
}
