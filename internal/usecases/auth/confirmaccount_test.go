package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

func TestConfirmAccountUseCase_Success(t *testing.T) {
	gw := new(mocks.UserGateway)

	userID := uuid.New()
	codeID := uuid.New()

	gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
	gw.On("GetValidConfirmationCode", mock.Anything, userID, "the-code", user.PurposeAccountConfirmation).
		Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
	gw.On("ActivateUser", mock.Anything, userID).Return(nil)
	gw.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(nil)

	uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, gw)

	err := uc.Execute(context.Background(), userID, "the-code")

	require.NoError(t, err)
	gw.AssertExpectations(t)
}

func TestConfirmAccountUseCase_AlreadyConfirmed(t *testing.T) {
	gw := new(mocks.UserGateway)

	userID := uuid.New()
	gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: true}, nil)

	uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, gw)

	err := uc.Execute(context.Background(), userID, "the-code")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.BE1008_ACCOUNT_ALREADY_CONFIRMED, *appErr.BusinessCode())
	gw.AssertNotCalled(t, "ActivateUser", mock.Anything, mock.Anything)
}

func TestConfirmAccountUseCase_InvalidCode(t *testing.T) {
	gw := new(mocks.UserGateway)

	userID := uuid.New()
	gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
	gw.On("GetValidConfirmationCode", mock.Anything, userID, "nope", user.PurposeAccountConfirmation).
		Return(nil, nil)

	uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, gw)

	err := uc.Execute(context.Background(), userID, "nope")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.BE1006_CONFIRMATION_CODE_INVALID, *appErr.BusinessCode())
}

func TestConfirmAccountUseCase_ExpiredCode(t *testing.T) {
	gw := new(mocks.UserGateway)

	userID := uuid.New()
	gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
	gw.On("GetValidConfirmationCode", mock.Anything, userID, "old", user.PurposeAccountConfirmation).
		Return(&user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}, nil)

	uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, gw)

	err := uc.Execute(context.Background(), userID, "old")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
	gw.AssertNotCalled(t, "ActivateUser", mock.Anything, mock.Anything)
}
