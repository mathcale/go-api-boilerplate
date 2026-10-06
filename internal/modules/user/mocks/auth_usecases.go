package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/usecases"
)

type SignInUseCase struct {
	mock.Mock
}

func (m *SignInUseCase) Execute(
	ctx context.Context,
	email, password string,
) (*usecases.TokenPair, error) {
	args := m.Called(ctx, email, password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*usecases.TokenPair), args.Error(1)
}

type SignUpUseCase struct {
	mock.Mock
}

func (m *SignUpUseCase) Execute(
	ctx context.Context,
	input usecases.SignUpInput,
) (*domain.User, error) {
	args := m.Called(ctx, input)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

type RefreshTokenUseCase struct {
	mock.Mock
}

func (m *RefreshTokenUseCase) Execute(
	ctx context.Context,
	refreshToken string,
) (*usecases.TokenPair, error) {
	args := m.Called(ctx, refreshToken)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*usecases.TokenPair), args.Error(1)
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

func (m *MeUseCase) Execute(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
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
