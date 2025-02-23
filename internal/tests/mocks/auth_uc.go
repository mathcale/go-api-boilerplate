package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
)

type SignInUseCase struct {
	mock.Mock
}

type SignUpUseCase struct {
	mock.Mock
}

type SignInGateway struct {
	mock.Mock
}

type SignUpGateway struct {
	mock.Mock
}

func (m *SignInUseCase) Execute(ctx context.Context, user domain.User) (accessToken, refreshToken *string, err error) {
	args := m.Called(ctx, user)

	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}

	if args.Get(1) == nil {
		return nil, nil, args.Error(2)
	}

	return args.Get(0).(*string), args.Get(1).(*string), args.Error(2)
}

func (m *SignUpUseCase) Execute(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *SignInGateway) GetUser(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *SignUpGateway) UserExists(ctx context.Context, email string) (*bool, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*bool), args.Error(1)
}

func (m *SignUpGateway) SaveUser(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
