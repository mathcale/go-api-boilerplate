package gateway

import (
	"context"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
)

type Gateway interface {
	UserExistsIncludingInactive(ctx context.Context, email string) (*bool, error)
	SaveUser(ctx context.Context, u domain.User, code domain.ConfirmationCode) error
	GetUserByEmailIncludingInactive(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	SaveConfirmationCode(ctx context.Context, code domain.ConfirmationCode) error
	GetValidConfirmationCode(
		ctx context.Context,
		userID uuid.UUID,
		code string,
		purpose domain.Purpose,
	) (*domain.ConfirmationCode, error)
	MarkConfirmationCodeUsed(ctx context.Context, codeID uuid.UUID) error
	SendConfirmationEmail(ctx context.Context, u domain.User, code domain.ConfirmationCode) error
	SendRecoveryEmail(ctx context.Context, u domain.User, code domain.ConfirmationCode) error
	SaveRefreshToken(ctx context.Context, token domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*domain.RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) error
	RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error
}
