package mocks

import "github.com/stretchr/testify/mock"

type BcryptPassword struct {
	mock.Mock
}

func (m *BcryptPassword) Hash(password string) (*string, error) {
	args := m.Called(password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*string), args.Error(1)
}

func (m *BcryptPassword) Verify(password string, hashedPassword string) error {
	args := m.Called(password, hashedPassword)
	return args.Error(0)
}
