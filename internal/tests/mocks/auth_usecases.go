package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type SignInUseCase struct {
	mock.Mock
}

func (m *SignInUseCase) Execute(
	ctx context.Context,
	email, password string,
) (*auth.TokenPair, error) {
	args := m.Called(ctx, email, password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*auth.TokenPair), args.Error(1)
}
