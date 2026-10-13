package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type SignUpUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	pw *mocks.Password
}

func (s *SignUpUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.pw = new(mocks.Password)
}

func (s *SignUpUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestSignUpUseCase(t *testing.T) {
	suite.Run(t, new(SignUpUseCaseTestSuite))
}

func (s *SignUpUseCaseTestSuite) TestExecute() {
	s.Run("should create an inactive user with hashed password and send confirmation", func() {
		hashed := fixtures.AuthBHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(fixtures.AuthBBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthBPassword).Return(&hashed, nil)
		s.gw.On(
			"SaveUser", mock.Anything, mock.AnythingOfType("user.User"),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On(
			"SendConfirmationEmail", mock.Anything, mock.AnythingOfType("user.User"),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInput())

		s.Require().NoError(err)
		s.Require().NotNil(created)
		s.Equal(fixtures.AuthBEmail, created.Email)
		s.False(created.Active)
		s.Equal(hashed, created.Password)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
	})

	s.Run("should still create the user even when sending the confirmation email fails", func() {
		hashed := fixtures.AuthBHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(fixtures.AuthBBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthBPassword).Return(&hashed, nil)
		s.gw.On("SaveUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("smtp down"))

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInput())

		s.Require().NoError(err)
		s.Require().NotNil(created)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should return a conflict error when the email is already registered", func() {
		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(fixtures.AuthBBoolPtr(true), nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInput())

		s.Require().Error(err)
		s.Nil(created)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ConflictKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1000_USER_ALREADY_EXISTS, *appErr.BusinessCode())
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
		s.gw.AssertNotCalled(s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error from checking if the user already exists", func() {
		gwErr := errors.New("db down")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(nil, gwErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInput())

		s.Require().ErrorIs(err, gwErr)
		s.Nil(created)
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
	})

	s.Run("should propagate the password hashing error without saving the user", func() {
		hashErr := errors.New("boom")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(fixtures.AuthBBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthBPassword).Return(nil, hashErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInput())

		s.Require().ErrorIs(err, hashErr)
		s.Nil(created)
		s.gw.AssertNotCalled(s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return a validation error when the domain input is invalid", func() {
		hashed := fixtures.AuthBHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(fixtures.AuthBBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthBPassword).Return(&hashed, nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInputWithoutName())

		s.Require().Error(err)
		s.Nil(created)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should propagate the save error and not send the confirmation email", func() {
		hashed := fixtures.AuthBHashedPassword
		saveErr := errors.New("db down")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthBEmail).
			Return(fixtures.AuthBBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthBPassword).Return(&hashed, nil)
		s.gw.On(
			"SaveUser", mock.Anything, mock.AnythingOfType("user.User"),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(saveErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthBSignUpInput())

		s.Require().ErrorIs(err, saveErr)
		s.Nil(created)
		s.gw.AssertNotCalled(
			s.T(), "SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything,
		)
	})
}
