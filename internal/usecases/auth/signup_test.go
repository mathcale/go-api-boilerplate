package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

func TestSignUpUseCase_Success(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	hashed := "hashed-password"
	falsy := false

	gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
	pw.On("Hash", "secret123").Return(&hashed, nil)
	gw.On("SaveUser", mock.Anything, mock.AnythingOfType("user.User"), mock.AnythingOfType("user.ConfirmationCode")).Return(nil)
	gw.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("user.User"), mock.AnythingOfType("user.ConfirmationCode")).Return(nil)

	uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, gw, pw)

	created, err := uc.Execute(context.Background(), auth.SignUpInput{
		Name:     "Jane",
		Surname:  "Doe",
		Email:    "jane@example.com",
		Password: "secret123",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "jane@example.com", created.Email)
	assert.False(t, created.Active)
	assert.Equal(t, hashed, created.Password)
	gw.AssertExpectations(t)
	pw.AssertExpectations(t)
}

func TestSignUpUseCase_AlreadyExists(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	truthy := true
	gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&truthy, nil)

	uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, gw, pw)

	created, err := uc.Execute(context.Background(), auth.SignUpInput{
		Name:     "Jane",
		Surname:  "Doe",
		Email:    "jane@example.com",
		Password: "secret123",
	})

	require.Error(t, err)
	assert.Nil(t, created)

	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.ConflictKind, appErr.Kind())
	require.NotNil(t, appErr.BusinessCode())
	assert.Equal(t, apperror.BE1000_USER_ALREADY_EXISTS, *appErr.BusinessCode())

	pw.AssertNotCalled(t, "Hash", mock.Anything)
	gw.AssertNotCalled(t, "SaveUser", mock.Anything, mock.Anything, mock.Anything)
}

func TestSignUpUseCase_HashErrorPropagates(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	falsy := false
	hashErr := errors.New("boom")

	gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
	pw.On("Hash", "secret123").Return(nil, hashErr)

	uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, gw, pw)

	created, err := uc.Execute(context.Background(), auth.SignUpInput{
		Name:     "Jane",
		Surname:  "Doe",
		Email:    "jane@example.com",
		Password: "secret123",
	})

	require.ErrorIs(t, err, hashErr)
	assert.Nil(t, created)
	gw.AssertNotCalled(t, "SaveUser", mock.Anything, mock.Anything, mock.Anything)
}

// Email delivery failures must not fail the sign-up: the account is created and
// the user can request a fresh code.
func TestSignUpUseCase_EmailFailureIsNonFatal(t *testing.T) {
	gw := new(mocks.UserGateway)
	pw := new(mocks.Password)

	hashed := "hashed-password"
	falsy := false

	gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
	pw.On("Hash", "secret123").Return(&hashed, nil)
	gw.On("SaveUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	gw.On("SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("smtp down"))

	uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, gw, pw)

	created, err := uc.Execute(context.Background(), auth.SignUpInput{
		Name:     "Jane",
		Surname:  "Doe",
		Email:    "jane@example.com",
		Password: "secret123",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	gw.AssertExpectations(t)
}
