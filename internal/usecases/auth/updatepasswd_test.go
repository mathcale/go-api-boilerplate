package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
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

func (s *UpdatePasswordUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestUpdatePasswordUseCase(t *testing.T) {
	suite.Run(t, new(UpdatePasswordUseCaseTestSuite))
}

func (s *UpdatePasswordUseCaseTestSuite) TestExecute() {
	s.Run("should update the password and mark the confirmation code as used", func() {
		userID := fixtures.AuthBNewID()
		code := fixtures.AuthBValidConfirmationCode()
		hashed := fixtures.AuthBNewHash

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthBCode,
			user.PurposePasswordRecovery,
		).Return(code, nil)
		s.pw.On("Hash", fixtures.AuthBNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(nil)
		s.gw.On("MarkConfirmationCodeUsed", mock.Anything, code.ID).Return(nil)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthBCode, fixtures.AuthBNewPassword,
		)

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
	})

	s.Run("should return a recovery code mismatch error when the code is invalid", func() {
		userID := fixtures.AuthBNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "bad",
			user.PurposePasswordRecovery,
		).Return(nil, nil)

		err := s.uc.Execute(context.Background(), userID, "bad", fixtures.AuthBNewPassword)

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
		s.gw.AssertNotCalled(s.T(), "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return an error when the confirmation code has expired", func() {
		userID := fixtures.AuthBNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "old",
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthBExpiredConfirmationCode(), nil)

		err := s.uc.Execute(context.Background(), userID, "old", fixtures.AuthBNewPassword)

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when fetching the confirmation code fails", func() {
		userID := fixtures.AuthBNewID()
		gwErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthBCode,
			user.PurposePasswordRecovery,
		).Return(nil, gwErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthBCode, fixtures.AuthBNewPassword,
		)

		s.Require().ErrorIs(err, gwErr)
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
	})

	s.Run("should propagate the error when hashing the new password fails", func() {
		userID := fixtures.AuthBNewID()
		hashErr := errors.New("hash failure")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthBCode,
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthBValidConfirmationCode(), nil)
		s.pw.On("Hash", fixtures.AuthBNewPassword).Return(nil, hashErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthBCode, fixtures.AuthBNewPassword,
		)

		s.Require().ErrorIs(err, hashErr)
		s.gw.AssertNotCalled(s.T(), "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when updating the password fails", func() {
		userID := fixtures.AuthBNewID()
		hashed := fixtures.AuthBNewHash
		updateErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthBCode,
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthBValidConfirmationCode(), nil)
		s.pw.On("Hash", fixtures.AuthBNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(updateErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthBCode, fixtures.AuthBNewPassword,
		)

		s.Require().ErrorIs(err, updateErr)
		s.gw.AssertNotCalled(s.T(), "MarkConfirmationCodeUsed", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when marking the confirmation code as used fails", func() {
		userID := fixtures.AuthBNewID()
		code := fixtures.AuthBValidConfirmationCode()
		hashed := fixtures.AuthBNewHash
		markErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthBCode,
			user.PurposePasswordRecovery,
		).Return(code, nil)
		s.pw.On("Hash", fixtures.AuthBNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePassword", mock.Anything, userID, hashed).Return(nil)
		s.gw.On("MarkConfirmationCodeUsed", mock.Anything, code.ID).Return(markErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthBCode, fixtures.AuthBNewPassword,
		)

		s.Require().ErrorIs(err, markErr)
	})
}
