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

func activeUser() *user.User {
	return &user.User{
		ID:       uuid.New(),
		Email:    "jane@example.com",
		Password: "stored-hash",
		Active:   true,
	}
}

func TestSignInUseCase_Success(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)
	j := new(mocks.JWTAuth)

	u := activeUser()
	access := "access-token"
	refresh := "refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
	pw.On("Verify", "secret123", "stored-hash").Return(nil)
	j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
		Return(&jwt.Token{Token: &access}, nil)
	j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
		Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
	gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
		Return(nil)

	uc := auth.NewSignInUseCase(mocks.NoopLogger{}, gw, pw, j)

	tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.Equal(t, access, tokens.AccessToken)
	assert.Equal(t, refresh, tokens.RefreshToken)
	gw.AssertExpectations(t)
	pw.AssertExpectations(t)
	j.AssertExpectations(t)
}

func TestSignInUseCase_SaveRefreshTokenFails(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)
	j := new(mocks.JWTAuth)

	u := activeUser()
	access := "access-token"
	refresh := "refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
	pw.On("Verify", "secret123", "stored-hash").Return(nil)
	j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
		Return(&jwt.Token{Token: &access}, nil)
	j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
		Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
	gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
		Return(errors.New("db down"))

	uc := auth.NewSignInUseCase(mocks.NoopLogger{}, gw, pw, j)

	tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

	require.Error(t, err)
	assert.Nil(t, tokens)
}

func TestSignInUseCase_UnknownUser(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)
	j := new(mocks.JWTAuth)

	gw.On("GetUserByEmailIncludingInactive", mock.Anything, "nobody@example.com").
		Return(nil, nil)

	uc := auth.NewSignInUseCase(mocks.NoopLogger{}, gw, pw, j)

	tokens, err := uc.Execute(context.Background(), "nobody@example.com", "secret123")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.UnauthorizedKind, appErr.Kind())
	require.NotNil(t, appErr.BusinessCode())
	assert.Equal(t, apperror.BE1010_INVALID_CREDENTIALS, *appErr.BusinessCode())
}

func TestSignInUseCase_WrongPassword(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)
	j := new(mocks.JWTAuth)

	u := activeUser()
	gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
	pw.On("Verify", "wrong", "stored-hash").Return(errors.New("mismatch"))

	uc := auth.NewSignInUseCase(mocks.NoopLogger{}, gw, pw, j)

	tokens, err := uc.Execute(context.Background(), "jane@example.com", "wrong")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.UnauthorizedKind, appErr.Kind())
	j.AssertNotCalled(t, "IssueAccessToken", mock.Anything)
}

func TestSignInUseCase_InactiveAccount(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)
	j := new(mocks.JWTAuth)

	u := activeUser()
	u.Active = false

	gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
	pw.On("Verify", "secret123", "stored-hash").Return(nil)

	uc := auth.NewSignInUseCase(mocks.NoopLogger{}, gw, pw, j)

	tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

	require.Error(t, err)
	assert.Nil(t, tokens)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.ForbiddenKind, appErr.Kind())
	require.NotNil(t, appErr.BusinessCode())
	assert.Equal(t, apperror.BE1009_ACCOUNT_NOT_CONFIRMED, *appErr.BusinessCode())
	j.AssertNotCalled(t, "IssueAccessToken", mock.Anything)
}
