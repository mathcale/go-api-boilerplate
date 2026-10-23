package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
	"go-api-boilerplate/internal/usecases/auth"
)

type ConfirmAccountUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	uc auth.ConfirmAccountUseCase
}

func (s *ConfirmAccountUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.uc = auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)
}

func (s *ConfirmAccountUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestConfirmAccountUseCase(t *testing.T) {
	suite.Run(t, new(ConfirmAccountUseCaseTestSuite))
}

func (s *ConfirmAccountUseCaseTestSuite) assertBusinessError(
	err error,
	kind apperror.Kind,
	code apperror.BusinessCode,
) {
	s.T().Helper()

	var appErr apperror.AppError
	s.Require().ErrorAs(err, &appErr)
	s.Equal(kind, appErr.Kind())
	s.Require().NotNil(appErr.BusinessCode())
	s.Equal(code, *appErr.BusinessCode())
}

func (s *ConfirmAccountUseCaseTestSuite) TestExecute() {
	s.Run("should consume the confirmation code and activate the user", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)
		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "the-code",
			user.PurposeAccountConfirmation,
		).Return(fixtures.AuthSessionValidConfirmationCode(codeID), nil)
		s.gw.On("ActivateUserWithCode", mock.Anything, userID, codeID).Return(true, nil)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should return a not found error when the user does not exist", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().Error(err)
		s.assertBusinessError(err, apperror.NotFoundKind, apperror.BE1001_USER_NOT_FOUND)
		s.gw.AssertNotCalled(s.T(), "ActivateUserWithCode", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return an error when the account is already confirmed", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, true), nil)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().Error(err)
		s.assertBusinessError(
			err, apperror.ConflictKind, apperror.BE1008_ACCOUNT_ALREADY_CONFIRMED,
		)
		s.gw.AssertNotCalled(s.T(), "ActivateUserWithCode", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return an error when the confirmation code is invalid", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)
		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "nope",
			user.PurposeAccountConfirmation,
		).Return(nil, nil)

		err := s.uc.Execute(context.Background(), userID, "nope")

		s.Require().Error(err)
		s.assertBusinessError(
			err, apperror.ValidationKind, apperror.BE1006_CONFIRMATION_CODE_INVALID,
		)
		s.gw.AssertNotCalled(s.T(), "ActivateUserWithCode", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return an error when the confirmation code has expired", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)
		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "old",
			user.PurposeAccountConfirmation,
		).Return(fixtures.AuthSessionExpiredConfirmationCode(uuid.New()), nil)

		err := s.uc.Execute(context.Background(), userID, "old")

		s.Require().Error(err)
		s.assertBusinessError(
			err, apperror.ValidationKind, apperror.BE1007_CONFIRMATION_CODE_EXPIRED,
		)
		s.gw.AssertNotCalled(s.T(), "ActivateUserWithCode", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when fetching the user fails", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, gwErr)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, gwErr)
	})

	s.Run("should propagate the error when fetching the confirmation code fails", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)
		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "the-code",
			user.PurposeAccountConfirmation,
		).Return(nil, gwErr)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, gwErr)
		s.gw.AssertNotCalled(s.T(), "ActivateUserWithCode", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return an invalid code error when the code was consumed concurrently", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)
		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "the-code",
			user.PurposeAccountConfirmation,
		).Return(fixtures.AuthSessionValidConfirmationCode(codeID), nil)
		s.gw.On("ActivateUserWithCode", mock.Anything, userID, codeID).Return(false, nil)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().Error(err)
		s.assertBusinessError(
			err, apperror.ValidationKind, apperror.BE1006_CONFIRMATION_CODE_INVALID,
		)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should propagate the error when consuming the code and activating fails", func() {
		userID := uuid.New()
		codeID := uuid.New()
		activateErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)
		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "the-code",
			user.PurposeAccountConfirmation,
		).Return(fixtures.AuthSessionValidConfirmationCode(codeID), nil)
		s.gw.On("ActivateUserWithCode", mock.Anything, userID, codeID).
			Return(false, activateErr)

		err := s.uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, activateErr)
	})
}
