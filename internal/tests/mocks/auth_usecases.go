package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
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

type SignUpUseCase struct {
	mock.Mock
}

func (m *SignUpUseCase) Execute(ctx context.Context, input auth.SignUpInput) (*user.User, error) {
	args := m.Called(ctx, input)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*user.User), args.Error(1)
}

type RefreshTokenUseCase struct {
	mock.Mock
}

func (m *RefreshTokenUseCase) Execute(
	ctx context.Context,
	refreshToken string,
) (*auth.TokenPair, error) {
	args := m.Called(ctx, refreshToken)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

type ConfirmAccountUseCase struct {
	mock.Mock
}

func (m *ConfirmAccountUseCase) Execute(ctx context.Context, userID uuid.UUID, code string) error {
	return m.Called(ctx, userID, code).Error(0)
}

type ResendConfirmationCodeUseCase struct {
	mock.Mock
}

func (m *ResendConfirmationCodeUseCase) Execute(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

type MeUseCase struct {
	mock.Mock
}

func (m *MeUseCase) Execute(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*user.User), args.Error(1)
}

type SetRecoveryCodeUseCase struct {
	mock.Mock
}

func (m *SetRecoveryCodeUseCase) Execute(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

type ValidateRecoveryCodeUseCase struct {
	mock.Mock
}

func (m *ValidateRecoveryCodeUseCase) Execute(ctx context.Context, userID uuid.UUID, code string) error {
	return m.Called(ctx, userID, code).Error(0)
}

type UpdatePasswordUseCase struct {
	mock.Mock
}

func (m *UpdatePasswordUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	code, newPassword string,
) error {
	return m.Called(ctx, userID, code, newPassword).Error(0)
}
