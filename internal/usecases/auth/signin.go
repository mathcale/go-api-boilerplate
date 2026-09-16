package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/gateway"
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type (
	SignInUseCase interface {
		Execute(ctx context.Context, email, password string) (*TokenPair, error)
	}

	signInUseCase struct {
		logger   logger.Logger
		gateway  gateway.User
		password bcrypt.Password
		jwtAuth  jwt.JWTAuth
	}
)

func NewSignInUseCase(
	l logger.Logger,
	gw gateway.User,
	pw bcrypt.Password,
	j jwt.JWTAuth,
) SignInUseCase {
	return &signInUseCase{
		logger:   l,
		gateway:  gw,
		password: pw,
		jwtAuth:  j,
	}
}

func (uc *signInUseCase) Execute(
	ctx context.Context,
	email, password string,
) (*TokenPair, error) {
	invalidCreds := apperror.BE1010_INVALID_CREDENTIALS

	target, err := uc.gateway.GetUserByEmailIncludingInactive(ctx, email)
	if err != nil {
		return nil, err
	}

	if target == nil {
		return nil, apperror.New(
			nil, "invalid credentials", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "signin", &invalidCreds, nil,
		)
	}

	if verifyErr := uc.password.Verify(password, target.Password); verifyErr != nil {
		uc.logger.Debug("password verification failed", map[string]interface{}{
			logFieldUserID: target.ID.String(),
		})

		return nil, apperror.New(
			nil, "invalid credentials", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "signin", &invalidCreds, nil,
		)
	}

	if !target.Active {
		notConfirmed := apperror.BE1009_ACCOUNT_NOT_CONFIRMED

		return nil, apperror.New(
			nil, "account not confirmed", apperror.ForbiddenKind,
			apperror.UseCaseOrigin, "signin", &notConfirmed, nil,
		)
	}

	roles := target.Roles
	if roles == nil {
		roles = []string{}
	}

	familyID := uuid.New()
	extraClaims := jwt.ExtraClaims{Roles: roles}

	accessToken, err := uc.jwtAuth.IssueAccessToken(jwt.IssueTokenParams{
		UserID:      target.ID.String(),
		ExtraClaims: extraClaims,
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtAuth.IssueRefreshToken(jwt.IssueTokenParams{
		UserID:      target.ID.String(),
		ExtraClaims: extraClaims,
		TokenID:     familyID.String(),
	})
	if err != nil {
		return nil, err
	}

	if refreshToken.ExpiresAt == nil {
		return nil, apperror.New(
			nil, "refresh token missing expiry", apperror.DependencyKind,
			apperror.UseCaseOrigin, "signin", nil, nil,
		)
	}

	record := user.NewRefreshTokenFamily(familyID, target.ID, *refreshToken.ExpiresAt)

	if err := uc.gateway.SaveRefreshToken(ctx, record); err != nil {
		return nil, err
	}

	pair := newTokenPair(accessToken, refreshToken)

	return &pair, nil
}
