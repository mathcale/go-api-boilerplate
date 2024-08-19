package mocks

import "github.com/stretchr/testify/mock"

type CounterUseCaseMock struct {
	mock.Mock
}

func (m *CounterUseCaseMock) Execute() (int, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return 0, args.Error(1)
	}

	return args.Int(0), args.Error(1)
}
