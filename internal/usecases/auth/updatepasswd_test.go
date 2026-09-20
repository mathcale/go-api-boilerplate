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

type UpdatePasswordUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	pw *mocks.Password
	uc auth.UpdatePasswordUseCase
}

func (s *UpdatePasswordUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.pw = new(mocks.Password)
	s.uc = auth.NewUpdatePasswordUseCase(mocks.NoopLogger{}, s.gw, s.pw)
}

func TestUpdatePasswordUseCase(t *testing.T) {
	suite.Run(t, new(UpdatePasswordUseCaseTestSuite))
}

func (s *UpdatePasswordUseCaseTestSuite) TestSuccess() {
	s.Run("should update the password and mark the confirmation code as used", func() {
		userID := uuid.New()
		codeID := uuid.New()
		hashed := "new-hash"

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", user.PurposePasswordRecovery).
			Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.pw.On("Hash", "new-password").Return(&hashed, nil)
		s.gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(nil)
		s.gw.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(nil)

		err := s.uc.Execute(context.Background(), userID, "code", "new-password")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
	})
}

func (s *UpdatePasswordUseCaseTestSuite) TestInvalidCode() {
	s.Run("should return a recovery code mismatch error when the code is invalid", func() {
		userID := uuid.New()
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "bad", user.PurposePasswordRecovery).
			Return(nil, nil)

		err := s.uc.Execute(context.Background(), userID, "bad", "new-password")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
		s.gw.AssertNotCalled(s.T(), "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})
}

func (s *UpdatePasswordUseCaseTestSuite) TestExpiredCode() {
	s.Run("should return an error when the confirmation code has expired", func() {
		userID := uuid.New()
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "old", user.PurposePasswordRecovery).
			Return(&user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}, nil)

		err := s.uc.Execute(context.Background(), userID, "old", "new-password")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})
}

func (s *UpdatePasswordUseCaseTestSuite) TestGetValidConfirmationCodeError() {
	s.Run("should propagate the error when fetching the confirmation code fails", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")
		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", user.PurposePasswordRecovery).
			Return(nil, gwErr)

		err := s.uc.Execute(context.Background(), userID, "code", "new-password")

		s.Require().ErrorIs(err, gwErr)
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
	})
}

func (s *UpdatePasswordUseCaseTestSuite) TestHashError() {
	s.Run("should propagate the error when hashing the new password fails", func() {
		userID := uuid.New()
		codeID := uuid.New()
		hashErr := errors.New("hash failure")

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", user.PurposePasswordRecovery).
			Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.pw.On("Hash", "new-password").Return(nil, hashErr)

		err := s.uc.Execute(context.Background(), userID, "code", "new-password")

		s.Require().ErrorIs(err, hashErr)
		s.gw.AssertNotCalled(s.T(), "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})
}

func (s *UpdatePasswordUseCaseTestSuite) TestUpdatePasswordError() {
	s.Run("should propagate the error when updating the password fails", func() {
		userID := uuid.New()
		codeID := uuid.New()
		hashed := "new-hash"
		updateErr := errors.New("db down")

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", user.PurposePasswordRecovery).
			Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.pw.On("Hash", "new-password").Return(&hashed, nil)
		s.gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(updateErr)

		err := s.uc.Execute(context.Background(), userID, "code", "new-password")

		s.Require().ErrorIs(err, updateErr)
		s.gw.AssertNotCalled(s.T(), "MarkConfirmationCodeUsed", mock.Anything, mock.Anything)
	})
}

func (s *UpdatePasswordUseCaseTestSuite) TestMarkConfirmationCodeUsedError() {
	s.Run("should propagate the error when marking the confirmation code as used fails", func() {
		userID := uuid.New()
		codeID := uuid.New()
		hashed := "new-hash"
		markErr := errors.New("db down")

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", user.PurposePasswordRecovery).
			Return(&user.ConfirmationCode{ID: codeID, ExpiresAt: time.Now().Add(time.Minute)}, nil)
		s.pw.On("Hash", "new-password").Return(&hashed, nil)
		s.gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(nil)
		s.gw.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(markErr)

		err := s.uc.Execute(context.Background(), userID, "code", "new-password")

		s.Require().ErrorIs(err, markErr)
	})
}
