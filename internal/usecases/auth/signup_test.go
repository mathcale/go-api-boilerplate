package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
	"go-api-boilerplate/internal/usecases/auth"
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
		hashed := fixtures.AuthCredsHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthCredsPassword).Return(&hashed, nil)
		s.gw.On(
			"SaveUser", mock.Anything, mock.AnythingOfType("user.User"),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On(
			"SendConfirmationEmail", mock.Anything, mock.AnythingOfType("user.User"),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInput())

		s.Require().NoError(err)
		s.Require().NotNil(created)
		s.Equal(fixtures.AuthCredsEmail, created.Email)
		s.False(created.Active)
		s.Equal(hashed, created.Password)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
	})

	s.Run("should normalise the email before checking existence and saving", func() {
		hashed := fixtures.AuthCredsHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthCredsPassword).Return(&hashed, nil)
		s.gw.On(
			"SaveUser", mock.Anything,
			mock.MatchedBy(func(u user.User) bool { return u.Email == fixtures.AuthCredsEmail }),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		_, err := uc.Execute(
			context.Background(), fixtures.AuthCredsSignUpInputWithMixedCaseEmail(),
		)

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should still create the user even when sending the confirmation email fails", func() {
		hashed := fixtures.AuthCredsHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthCredsPassword).Return(&hashed, nil)
		s.gw.On("SaveUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("smtp down"))

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInput())

		s.Require().NoError(err)
		s.Require().NotNil(created)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should return a conflict error when the email is already registered", func() {
		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(true), nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInput())

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

	s.Run("should return a conflict when the email only differs by casing", func() {
		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(true), nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		_, err := uc.Execute(
			context.Background(), fixtures.AuthCredsSignUpInputWithMixedCaseEmail(),
		)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ConflictKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1000_USER_ALREADY_EXISTS, *appErr.BusinessCode())
	})

	s.Run("should reject invalid raw passwords before hashing", func() {
		for _, pw := range []string{
			fixtures.AuthCredsShortPassword, "        ", fixtures.AuthCredsLongPassword(),
		} {
			s.SetupTest()
			s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
				Return(fixtures.AuthCredsBoolPtr(false), nil)

			uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

			created, err := uc.Execute(
				context.Background(), fixtures.AuthCredsSignUpInputWithPassword(pw),
			)

			s.Nil(created)

			var appErr apperror.AppError
			s.Require().ErrorAs(err, &appErr)
			s.Equal(apperror.ValidationKind, appErr.Kind())
			s.Require().NotNil(appErr.BusinessCode())
			s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
			s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
			s.gw.AssertNotCalled(
				s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything,
			)
		}
	})

	s.Run("should propagate the error from checking if the user already exists", func() {
		gwErr := errors.New("db down")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(nil, gwErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInput())

		s.Require().ErrorIs(err, gwErr)
		s.Nil(created)
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
	})

	s.Run("should propagate the password hashing error without saving the user", func() {
		hashErr := errors.New("boom")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthCredsPassword).Return(nil, hashErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInput())

		s.Require().ErrorIs(err, hashErr)
		s.Nil(created)
		s.gw.AssertNotCalled(s.T(), "SaveUser", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return a validation error when the domain input is invalid", func() {
		hashed := fixtures.AuthCredsHashedPassword

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthCredsPassword).Return(&hashed, nil)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInputWithoutName())

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
		hashed := fixtures.AuthCredsHashedPassword
		saveErr := errors.New("db down")

		s.gw.On("UserExistsIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsBoolPtr(false), nil)
		s.pw.On("Hash", fixtures.AuthCredsPassword).Return(&hashed, nil)
		s.gw.On(
			"SaveUser", mock.Anything, mock.AnythingOfType("user.User"),
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(saveErr)

		uc := auth.NewSignUpUseCase(mocks.NoopLogger{}, s.gw, s.pw)

		created, err := uc.Execute(context.Background(), fixtures.AuthCredsSignUpInput())

		s.Require().ErrorIs(err, saveErr)
		s.Nil(created)
		s.gw.AssertNotCalled(
			s.T(), "SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything,
		)
	})
}
