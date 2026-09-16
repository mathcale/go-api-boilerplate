package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
)

type UserGateway struct {
	mock.Mock
}

func (m *UserGateway) UserExistsIncludingInactive(
	ctx context.Context,
	email string,
) (*bool, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*bool), args.Error(1)
}

func (m *UserGateway) SaveUser(
	ctx context.Context,
	u user.User,
	code user.ConfirmationCode,
) error {
	return m.Called(ctx, u, code).Error(0)
}

func (m *UserGateway) GetUserByEmailIncludingInactive(
	ctx context.Context,
	email string,
) (*user.User, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*user.User), args.Error(1)
}

func (m *UserGateway) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*user.User), args.Error(1)
}

func (m *UserGateway) UpdatePassword(
	ctx context.Context,
	userID uuid.UUID,
	hashedPassword string,
) error {
	return m.Called(ctx, userID, hashedPassword).Error(0)
}

func (m *UserGateway) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *UserGateway) SaveConfirmationCode(
	ctx context.Context,
	code user.ConfirmationCode,
) error {
	return m.Called(ctx, code).Error(0)
}

func (m *UserGateway) GetValidConfirmationCode(
	ctx context.Context,
	userID uuid.UUID,
	code string,
	purpose user.Purpose,
) (*user.ConfirmationCode, error) {
	args := m.Called(ctx, userID, code, purpose)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*user.ConfirmationCode), args.Error(1)
}

func (m *UserGateway) MarkConfirmationCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	return m.Called(ctx, codeID).Error(0)
}

func (m *UserGateway) SendConfirmationEmail(
	ctx context.Context,
	u user.User,
	code user.ConfirmationCode,
) error {
	return m.Called(ctx, u, code).Error(0)
}

func (m *UserGateway) SendRecoveryEmail(
	ctx context.Context,
	u user.User,
	code user.ConfirmationCode,
) error {
	return m.Called(ctx, u, code).Error(0)
}

func (m *UserGateway) SaveRefreshToken(ctx context.Context, t user.RefreshToken) error {
	return m.Called(ctx, t).Error(0)
}

func (m *UserGateway) GetRefreshToken(
	ctx context.Context,
	tokenID uuid.UUID,
) (*user.RefreshToken, error) {
	args := m.Called(ctx, tokenID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*user.RefreshToken), args.Error(1)
}

func (m *UserGateway) MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	return m.Called(ctx, tokenID).Error(0)
}

func (m *UserGateway) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	return m.Called(ctx, familyID).Error(0)
}
