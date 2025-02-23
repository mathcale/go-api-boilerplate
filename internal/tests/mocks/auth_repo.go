package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Exists(ctx context.Context, email string) (*bool, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*bool), args.Error(1)
}

func (m *UserRepositoryMock) Get(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepositoryMock) Save(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
