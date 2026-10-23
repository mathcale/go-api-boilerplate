package gateway

import (
	"context"

	"github.com/google/uuid"

	"go-api-boilerplate/internal/domain/user"
)

type User interface {
	UserExistsIncludingInactive(ctx context.Context, email string) (*bool, error)
	SaveUser(ctx context.Context, u user.User, code user.ConfirmationCode) error
	GetUserByEmailIncludingInactive(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error)
	GetUserAuthState(ctx context.Context, userID uuid.UUID) (*user.AuthState, error)
	UpdatePasswordWithCode(
		ctx context.Context,
		userID uuid.UUID,
		codeID uuid.UUID,
		hashedPassword string,
	) (bool, error)
	ActivateUserWithCode(ctx context.Context, userID uuid.UUID, codeID uuid.UUID) (bool, error)
	SaveConfirmationCode(ctx context.Context, code user.ConfirmationCode) error
	GetValidConfirmationCode(
		ctx context.Context,
		userID uuid.UUID,
		code string,
		purpose user.Purpose,
	) (*user.ConfirmationCode, error)
	SendConfirmationEmail(ctx context.Context, u user.User, code user.ConfirmationCode) error
	SendRecoveryEmail(ctx context.Context, u user.User, code user.ConfirmationCode) error
	SaveRefreshToken(ctx context.Context, token user.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*user.RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) (bool, error)
	RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}
