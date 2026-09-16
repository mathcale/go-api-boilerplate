package mocks

import "github.com/stretchr/testify/mock"

type Password struct {
	mock.Mock
}

func (m *Password) Hash(password string) (*string, error) {
	args := m.Called(password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*string), args.Error(1)
}

func (m *Password) Verify(password, hashedPassword string) error {
	return m.Called(password, hashedPassword).Error(0)
}
