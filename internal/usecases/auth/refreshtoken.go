package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/gateway"
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type RefreshTokenUseCase interface {
	Execute(ctx context.Context, refreshToken string) (*TokenPair, error)
}

type refreshTokenUseCase struct {
	logger  logger.Logger
	jwtAuth jwt.JWTAuth
	gateway gateway.User
}

func NewRefreshTokenUseCase(l logger.Logger, j jwt.JWTAuth, gw gateway.User) RefreshTokenUseCase {
	return &refreshTokenUseCase{
		logger:  l,
		jwtAuth: j,
		gateway: gw,
	}
}

func (uc *refreshTokenUseCase) Execute(
	ctx context.Context,
	refreshToken string,
) (*TokenPair, error) {
	stored, tokenVersion, err := uc.resolveStoredToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if stored.IsUsed() {
		return nil, uc.handleReuse(ctx, *stored)
	}

	if markErr := uc.gateway.MarkRefreshTokenUsed(ctx, stored.ID); markErr != nil {
		return nil, markErr
	}

	owner, err := uc.resolveOwner(ctx, *stored)
	if err != nil {
		return nil, err
	}

	if tokenVersion != owner.TokenVersion {
		return nil, uc.handleStaleVersion(ctx, *stored)
	}

	return uc.rotate(ctx, *owner, *stored)
}

func (uc *refreshTokenUseCase) resolveStoredToken(
	ctx context.Context,
	refreshToken string,
) (*user.RefreshToken, int, error) {
	unauthorized := apperror.BE0003_UNAUTHORIZED

	verified, err := uc.jwtAuth.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, 0, apperror.New(
			err, "invalid refresh token", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	if verified.ID == "" {
		return nil, 0, apperror.New(
			errors.New("missing jti"), "invalid refresh token", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	tokenID, err := uuid.Parse(verified.ID)
	if err != nil {
		return nil, 0, apperror.New(
			err, "invalid refresh token id", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	stored, err := uc.gateway.GetRefreshToken(ctx, tokenID)
	if err != nil {
		return nil, 0, err
	}

	if stored == nil || stored.IsRevoked() {
		return nil, 0, apperror.New(
			nil, "refresh token unknown or revoked", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	return stored, verified.TokenVersion, nil
}

func (uc *refreshTokenUseCase) resolveOwner(
	ctx context.Context,
	stored user.RefreshToken,
) (*user.User, error) {
	owner, err := uc.gateway.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	if owner == nil || !owner.Active {
		unauthorized := apperror.BE0003_UNAUTHORIZED

		return nil, apperror.New(
			nil, "refresh token owner not found or inactive", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	return owner, nil
}

func (uc *refreshTokenUseCase) rotate(
	ctx context.Context,
	owner user.User,
	stored user.RefreshToken,
) (*TokenPair, error) {
	newFamilyMemberID := uuid.New()

	roles := owner.Roles
	if roles == nil {
		roles = []string{}
	}

	extraClaims := jwt.ExtraClaims{Roles: roles}

	newAccessToken, err := uc.jwtAuth.IssueAccessToken(jwt.IssueTokenParams{
		UserID:       owner.ID.String(),
		ExtraClaims:  extraClaims,
		TokenVersion: owner.TokenVersion,
	})
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := uc.jwtAuth.IssueRefreshToken(jwt.IssueTokenParams{
		UserID:       owner.ID.String(),
		ExtraClaims:  extraClaims,
		TokenID:      newFamilyMemberID.String(),
		TokenVersion: owner.TokenVersion,
	})
	if err != nil {
		return nil, err
	}

	if newRefreshToken.ExpiresAt == nil {
		return nil, apperror.New(
			nil, "refresh token missing expiry", apperror.DependencyKind,
			apperror.UseCaseOrigin, "refresh_token", nil, nil,
		)
	}

	rotated := user.RotateRefreshToken(stored, newFamilyMemberID, *newRefreshToken.ExpiresAt)

	if err := uc.gateway.SaveRefreshToken(ctx, rotated); err != nil {
		return nil, err
	}

	pair := newTokenPair(newAccessToken, newRefreshToken)

	return &pair, nil
}

func (uc *refreshTokenUseCase) handleReuse(ctx context.Context, stored user.RefreshToken) error {
	unauthorized := apperror.BE0003_UNAUTHORIZED

	if revokeErr := uc.gateway.RevokeRefreshTokenFamily(ctx, stored.FamilyID); revokeErr != nil {
		uc.logger.Error(
			"failed to revoke refresh token family after reuse",
			revokeErr,
			"family_id", stored.FamilyID.String(),
		)

		return apperror.New(
			revokeErr, "refresh token reuse detected", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	uc.logger.Warn(
		"refresh token reuse detected; family revoked",
		logFieldUserID, stored.UserID.String(),
		"family_id", stored.FamilyID.String(),
		"business_code", apperror.BE1011_REFRESH_TOKEN_REUSE_DETECTED,
	)

	return apperror.New(
		nil, "refresh token reuse detected", apperror.UnauthorizedKind,
		apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
	)
}

func (uc *refreshTokenUseCase) handleStaleVersion(
	ctx context.Context,
	stored user.RefreshToken,
) error {
	unauthorized := apperror.BE0003_UNAUTHORIZED

	if revokeErr := uc.gateway.RevokeRefreshTokenFamily(ctx, stored.FamilyID); revokeErr != nil {
		uc.logger.Error(
			"failed to revoke refresh token family after stale token version",
			revokeErr,
			"family_id", stored.FamilyID.String(),
		)

		return apperror.New(
			revokeErr, "refresh token version is stale", apperror.UnauthorizedKind,
			apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
		)
	}

	uc.logger.Warn(
		"refresh token version is stale; family revoked",
		logFieldUserID, stored.UserID.String(),
		"family_id", stored.FamilyID.String(),
	)

	return apperror.New(
		nil, "refresh token version is stale", apperror.UnauthorizedKind,
		apperror.UseCaseOrigin, "refresh_token", &unauthorized, nil,
	)
}
