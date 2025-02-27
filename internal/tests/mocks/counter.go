package mocks

import "github.com/stretchr/testify/mock"

type CounterUseCase struct {
	mock.Mock
}

func (m *CounterUseCase) Execute(limit int) (int, error) {
	args := m.Called(limit)

	if args.Get(0) == nil {
		return 0, args.Error(1)
	}

	return args.Int(0), args.Error(1)
}
