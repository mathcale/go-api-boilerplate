package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
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

func TestSignUpUseCase(t *testing.T) {
	suite.Run(t, new(SignUpUseCaseTestSuite))
}

func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_Success() {
	s.Run("should create a new inactive user with hashed password and send confirmation email", func() {
		hashed := "hashed-password"
		falsy := false

		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
		s.pw.On("Hash", "secret123").Return(&hashed, nil)
		s.gw.On("SaveUser", mock.Anything, mock.AnythingOfType("user.User"), mock.AnythingOfType("user.ConfirmationCode")).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, mock.AnythingOfType("user.User"), mock.AnythingOfType("user.ConfirmationCode")).Return(nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "Jane",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

		s.Require().NoError(err)
		s.Require().NotNil(created)
		s.Equal("jane@example.com", created.Email)
		s.False(created.Active)
		s.Equal(hashed, created.Password)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
	})
}

func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_AlreadyExists() {
	s.Run("should return a conflict error when the email is already registered", func() {
		truthy := true
		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&truthy, nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "Jane",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

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
}

func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_HashErrorPropagates() {
	s.Run("should propagate the password hashing error without saving the user", func() {
		falsy := false
		hashErr := errors.New("boom")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
		s.pw.On("Hash", "secret123").Return(nil, hashErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "Jane",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

		s.Require().ErrorIs(err, hashErr)
		s.Nil(created)
		s.gw.AssertNotCalled(s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything)
	})
}

// Email delivery failures must not fail the sign-up: the account is created and
// the user can request a fresh code.
func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_EmailFailureIsNonFatal() {
	s.Run("should still create the user even when sending the confirmation email fails", func() {
		hashed := "hashed-password"
		falsy := false

		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
		s.pw.On("Hash", "secret123").Return(&hashed, nil)
		s.gw.On("SaveUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("smtp down"))

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "Jane",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

		s.Require().NoError(err)
		s.Require().NotNil(created)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_UserExistsCheckErrorPropagates() {
	s.Run("should propagate the error from checking if the user already exists", func() {
		gwErr := errors.New("db down")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(nil, gwErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "Jane",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

		s.Require().ErrorIs(err, gwErr)
		s.Nil(created)
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
	})
}

func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_InvalidDomainInput() {
	s.Run("should return a validation error when the domain input is invalid", func() {
		falsy := false
		hashed := "hashed-password"

		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
		s.pw.On("Hash", "secret123").Return(&hashed, nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

		s.Require().Error(err)
		s.Nil(created)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything)
	})
}

func (s *SignUpUseCaseTestSuite) TestSignUpUseCase_SaveUserErrorPropagates() {
	s.Run("should propagate the error when saving the user fails and not send the confirmation email", func() {
		falsy := false
		hashed := "hashed-password"
		saveErr := errors.New("db down")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&falsy, nil)
		s.pw.On("Hash", "secret123").Return(&hashed, nil)
		s.gw.On("SaveUser", mock.Anything, mock.AnythingOfType("user.User"), mock.AnythingOfType("user.ConfirmationCode")).
			Return(saveErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), auth.SignUpInput{
			Name:     "Jane",
			Surname:  "Doe",
			Email:    "jane@example.com",
			Password: "secret123",
		})

		s.Require().ErrorIs(err, saveErr)
		s.Nil(created)
		s.gw.AssertNotCalled(s.T(), "SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything)
	})
}
