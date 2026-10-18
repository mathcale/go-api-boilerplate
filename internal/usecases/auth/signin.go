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
		limiter  gateway.AttemptLimiter

		dummyHash string
	}
)

func NewSignInUseCase(
	l logger.Logger,
	gw gateway.User,
	pw bcrypt.Password,
	j jwt.JWTAuth,
	limiter gateway.AttemptLimiter,
) (SignInUseCase, error) {
	dummyHash, err := pw.Hash(uuid.NewString())
	if err != nil {
		return nil, err
	}

	return &signInUseCase{
		logger:    l,
		gateway:   gw,
		password:  pw,
		jwtAuth:   j,
		limiter:   limiter,
		dummyHash: *dummyHash,
	}, nil
}

func (uc *signInUseCase) Execute(
	ctx context.Context,
	email, password string,
) (*TokenPair, error) {
	invalidCreds := apperror.BE1010_INVALID_CREDENTIALS

	email = user.NormalizeEmail(email)

	if !uc.limiter.Allow(email) {
		rateLimited := apperror.BE0005_RATE_LIMITED

		return nil, apperror.New(
			nil, "too many sign-in attempts", apperror.TooManyRequests,
			apperror.UseCaseOrigin, "signin", &rateLimited, nil,
		)
	}

	target, err := uc.gateway.GetUserByEmailIncludingInactive(ctx, email)
	if err != nil {
		return nil, err
	}

	if target == nil {
		_ = uc.password.Verify(password, uc.dummyHash)

		return nil, apperror.New(
			nil, "invalid credentials", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "signin", &invalidCreds, nil,
		)
	}

	if verifyErr := uc.password.Verify(password, target.Password); verifyErr != nil {
		uc.logger.Debug("password verification failed", logFieldUserID, target.ID.String())

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

	return uc.issueTokenPair(ctx, target)
}

func (uc *signInUseCase) issueTokenPair(
	ctx context.Context,
	target *user.User,
) (*TokenPair, error) {
	roles := target.Roles
	if roles == nil {
		roles = []string{}
	}

	familyID := uuid.New()
	extraClaims := jwt.ExtraClaims{Roles: roles}

	accessToken, err := uc.jwtAuth.IssueAccessToken(jwt.IssueTokenParams{
		UserID:       target.ID.String(),
		ExtraClaims:  extraClaims,
		TokenVersion: target.TokenVersion,
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtAuth.IssueRefreshToken(jwt.IssueTokenParams{
		UserID:       target.ID.String(),
		ExtraClaims:  extraClaims,
		TokenID:      familyID.String(),
		TokenVersion: target.TokenVersion,
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
