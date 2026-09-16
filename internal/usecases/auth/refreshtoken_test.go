package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

func storedRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	return &user.RefreshToken{
		ID:        familyID,
		UserID:    userID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

func TestRefreshTokenUseCase_Success(t *testing.T) {
	gw := new(mocks.UserGateway)
	j := new(mocks.JWTAuth)

	userID := uuid.New()
	familyID := uuid.New()
	stored := storedRefreshToken(userID, familyID)

	access := "new-access-token"
	refresh := "new-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	j.On("VerifyRefreshToken", "old-refresh-token").
		Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
	gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
	gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(nil)
	j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
		Return(&jwt.Token{Token: &access}, nil)
	j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
		Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
	gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
		Return(nil)

	uc := auth.NewRefreshTokenUseCase(mocks.NoopLogger{}, j, gw)

	tokens, err := uc.Execute(context.Background(), "old-refresh-token")

	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.Equal(t, access, tokens.AccessToken)
	assert.Equal(t, refresh, tokens.RefreshToken)
	gw.AssertExpectations(t)
	j.AssertExpectations(t)
}

func TestRefreshTokenUseCase_InvalidJWT(t *testing.T) {
	gw := new(mocks.UserGateway)
	j := new(mocks.JWTAuth)

	j.On("VerifyRefreshToken", "bad-token").Return(nil, errors.New("invalid"))

	uc := auth.NewRefreshTokenUseCase(mocks.NoopLogger{}, j, gw)

	tokens, err := uc.Execute(context.Background(), "bad-token")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.UnauthorizedKind, appErr.Kind())
	gw.AssertNotCalled(t, "GetRefreshToken", mock.Anything, mock.Anything)
}

func TestRefreshTokenUseCase_UnknownTokenID(t *testing.T) {
	gw := new(mocks.UserGateway)
	j := new(mocks.JWTAuth)

	userID := uuid.New()
	familyID := uuid.New()

	j.On("VerifyRefreshToken", "old-refresh-token").
		Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
	gw.On("GetRefreshToken", mock.Anything, familyID).Return(nil, nil)

	uc := auth.NewRefreshTokenUseCase(mocks.NoopLogger{}, j, gw)

	tokens, err := uc.Execute(context.Background(), "old-refresh-token")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.UnauthorizedKind, appErr.Kind())
	j.AssertNotCalled(t, "IssueAccessToken", mock.Anything)
}

func TestRefreshTokenUseCase_RevokedFamily(t *testing.T) {
	gw := new(mocks.UserGateway)
	j := new(mocks.JWTAuth)

	userID := uuid.New()
	familyID := uuid.New()
	stored := storedRefreshToken(userID, familyID)
	revokedAt := time.Now()
	stored.RevokedAt = &revokedAt

	j.On("VerifyRefreshToken", "old-refresh-token").
		Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
	gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)

	uc := auth.NewRefreshTokenUseCase(mocks.NoopLogger{}, j, gw)

	tokens, err := uc.Execute(context.Background(), "old-refresh-token")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.UnauthorizedKind, appErr.Kind())
	gw.AssertNotCalled(t, "RevokeRefreshTokenFamily", mock.Anything, mock.Anything)
	j.AssertNotCalled(t, "IssueAccessToken", mock.Anything)
}

func TestRefreshTokenUseCase_ReuseDetected(t *testing.T) {
	gw := new(mocks.UserGateway)
	j := new(mocks.JWTAuth)
	log := new(mocks.Logger)

	userID := uuid.New()
	familyID := uuid.New()
	stored := storedRefreshToken(userID, familyID)
	usedAt := time.Now()
	stored.UsedAt = &usedAt

	j.On("VerifyRefreshToken", "old-refresh-token").
		Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
	gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
	gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(nil)
	log.On("Warn", mock.Anything, mock.MatchedBy(func(tags map[string]interface{}) bool {
		return tags["business_code"] == apperror.BE1011_REFRESH_TOKEN_REUSE_DETECTED
	})).Return()

	uc := auth.NewRefreshTokenUseCase(log, j, gw)

	tokens, err := uc.Execute(context.Background(), "old-refresh-token")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.UnauthorizedKind, appErr.Kind())
	require.NotNil(t, appErr.BusinessCode())
	// The client-facing code stays generic: reuse detection must not be
	// distinguishable from any other unauthorized refresh via the API response.
	assert.Equal(t, apperror.BE0003_UNAUTHORIZED, *appErr.BusinessCode())
	gw.AssertExpectations(t)
	log.AssertExpectations(t)
	j.AssertNotCalled(t, "IssueAccessToken", mock.Anything)
	j.AssertNotCalled(t, "IssueRefreshToken", mock.Anything)
}
