package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type ConfirmAccountUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
}

func (s *ConfirmAccountUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
}

func TestConfirmAccountUseCase(t *testing.T) {
	suite.Run(t, new(ConfirmAccountUseCaseTestSuite))
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_Success() {
	s.Run("should activate the user and mark the confirmation code as used", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "the-code", user.PurposeAccountConfirmation).
			Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.gw.On("ActivateUser", mock.Anything, userID).Return(nil)
		s.gw.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(nil)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_AlreadyConfirmed() {
	s.Run("should return an error when the account is already confirmed", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: true}, nil)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1008_ACCOUNT_ALREADY_CONFIRMED, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "ActivateUser", mock.Anything, mock.Anything)
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_InvalidCode() {
	s.Run("should return an error when the confirmation code is invalid", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "nope", user.PurposeAccountConfirmation).
			Return(nil, nil)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "nope")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1006_CONFIRMATION_CODE_INVALID, *appErr.BusinessCode())
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_ExpiredCode() {
	s.Run("should return an error when the confirmation code has expired", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "old", user.PurposeAccountConfirmation).
			Return(&user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}, nil)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "old")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "ActivateUser", mock.Anything, mock.Anything)
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_UserNotFound() {
	s.Run("should return a not found error when the user does not exist", func() {
		userID := uuid.New()
		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.NotFoundKind, appErr.Kind())
		s.Equal(apperror.BE1001_USER_NOT_FOUND, *appErr.BusinessCode())
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_GetUserByIDError() {
	s.Run("should propagate the error when fetching the user fails", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, gwErr)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, gwErr)
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_GetValidConfirmationCodeError() {
	s.Run("should propagate the error when fetching the confirmation code fails", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "the-code", user.PurposeAccountConfirmation).
			Return(nil, gwErr)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, gwErr)
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_ActivateUserError() {
	s.Run("should propagate the error when activating the user fails and not mark the code as used", func() {
		userID := uuid.New()
		activateErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "the-code", user.PurposeAccountConfirmation).
			Return(&user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.gw.On("ActivateUser", mock.Anything, userID).Return(activateErr)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, activateErr)
		s.gw.AssertNotCalled(s.T(), "MarkConfirmationCodeUsed", mock.Anything, mock.Anything)
	})
}

func (s *ConfirmAccountUseCaseTestSuite) TestConfirmAccountUseCase_MarkConfirmationCodeUsedError() {
	s.Run("should propagate the error when marking the confirmation code as used fails", func() {
		userID := uuid.New()
		codeID := uuid.New()
		markErr := errors.New("db down")

		s.gw.On("GetUserByID", mock.Anything, userID).Return(&user.User{ID: userID, Active: false}, nil)
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "the-code", user.PurposeAccountConfirmation).
			Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.gw.On("ActivateUser", mock.Anything, userID).Return(nil)
		s.gw.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(markErr)

		uc := auth.NewConfirmAccountUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().ErrorIs(err, markErr)
	})
}
