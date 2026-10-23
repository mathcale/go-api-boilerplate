package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"go-api-boilerplate/internal/infra/database/models"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) ExistsIncludingInactive(ctx context.Context, email string) (*bool, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*bool), args.Error(1)
}

func (m *UserRepository) GetIncludingInactive(
	ctx context.Context,
	email string,
) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) Save(
	ctx context.Context,
	u models.User,
	code models.ConfirmationCode,
) error {
	return m.Called(ctx, u, code).Error(0)
}

func (m *UserRepository) GetAuthState(
	ctx context.Context,
	userID uuid.UUID,
) (*models.UserAuthState, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.UserAuthState), args.Error(1)
}

func (m *UserRepository) UpdatePasswordWithCode(
	ctx context.Context,
	userID uuid.UUID,
	codeID uuid.UUID,
	hashedPassword string,
) (bool, error) {
	args := m.Called(ctx, userID, codeID, hashedPassword)

	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) ActivateUserWithCode(
	ctx context.Context,
	userID uuid.UUID,
	codeID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, userID, codeID)

	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) SaveConfirmationCode(
	ctx context.Context,
	code models.ConfirmationCode,
) error {
	return m.Called(ctx, code).Error(0)
}

func (m *UserRepository) GetConfirmationCode(
	ctx context.Context,
	userID uuid.UUID,
	code string,
	purpose string,
) (*models.ConfirmationCode, error) {
	args := m.Called(ctx, userID, code, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.ConfirmationCode), args.Error(1)
}

func (m *UserRepository) SaveRefreshToken(ctx context.Context, token models.RefreshToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *UserRepository) GetRefreshToken(
	ctx context.Context,
	tokenID uuid.UUID,
) (*models.RefreshToken, error) {
	args := m.Called(ctx, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *UserRepository) MarkRefreshTokenUsed(
	ctx context.Context,
	tokenID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, tokenID)

	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	return m.Called(ctx, familyID).Error(0)
}

func (m *UserRepository) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
