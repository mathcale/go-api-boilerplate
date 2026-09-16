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

func TestUpdatePasswordUseCase_Success(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	userID := uuid.New()
	codeID := uuid.New()
	hashed := "new-hash"

	gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", user.PurposePasswordRecovery).
		Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
	pw.On("Hash", "new-password").Return(&hashed, nil)
	gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(nil)
	gw.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(nil)

	uc := auth.NewUpdatePasswordUseCase(mocks.NoopLogger{}, gw, pw)

	err := uc.Execute(context.Background(), userID, "code", "new-password")

	require.NoError(t, err)
	gw.AssertExpectations(t)
	pw.AssertExpectations(t)
}

func TestUpdatePasswordUseCase_InvalidCode(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	userID := uuid.New()
	gw.On("GetValidConfirmationCode", mock.Anything, userID, "bad", user.PurposePasswordRecovery).
		Return(nil, nil)

	uc := auth.NewUpdatePasswordUseCase(mocks.NoopLogger{}, gw, pw)

	err := uc.Execute(context.Background(), userID, "bad", "new-password")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
	pw.AssertNotCalled(t, "Hash", mock.Anything)
	gw.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdatePasswordUseCase_ExpiredCode(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	userID := uuid.New()
	gw.On("GetValidConfirmationCode", mock.Anything, userID, "old", user.PurposePasswordRecovery).
		Return(&user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}, nil)

	uc := auth.NewUpdatePasswordUseCase(mocks.NoopLogger{}, gw, pw)

	err := uc.Execute(context.Background(), userID, "old", "new-password")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
	gw.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
}
