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
	s.Run("should consume the code, update the password and revoke sessions", func() {
		userID := fixtures.AuthCredsNewID()
		code := fixtures.AuthCredsValidConfirmationCode()
		hashed := fixtures.AuthCredsNewHash

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(code, nil)
		s.pw.On("Hash", fixtures.AuthCredsNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePasswordWithCode", mock.Anything, userID, code.ID, hashed).
			Return(true, nil)
		s.gw.On("RevokeUserRefreshTokens", mock.Anything, userID).Return(nil)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsNewPassword,
		)

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
	})

	s.Run("should return a recovery code mismatch error when the code is invalid", func() {
		userID := fixtures.AuthCredsNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "bad",
			user.PurposePasswordRecovery,
		).Return(nil, nil)

		err := s.uc.Execute(context.Background(), userID, "bad", fixtures.AuthCredsNewPassword)

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
		s.gw.AssertNotCalled(s.T(), "UpdatePasswordWithCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return an error when the confirmation code has expired", func() {
		userID := fixtures.AuthCredsNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "old",
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthCredsExpiredConfirmationCode(), nil)

		err := s.uc.Execute(context.Background(), userID, "old", fixtures.AuthCredsNewPassword)

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "UpdatePasswordWithCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when fetching the confirmation code fails", func() {
		userID := fixtures.AuthCredsNewID()
		gwErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(nil, gwErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsNewPassword,
		)

		s.Require().ErrorIs(err, gwErr)
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
	})

	s.Run("should reject an invalid raw password before hashing", func() {
		userID := fixtures.AuthCredsNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthCredsValidConfirmationCode(), nil)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsLongPassword(),
		)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
		s.pw.AssertNotCalled(s.T(), "Hash", mock.Anything)
		s.gw.AssertNotCalled(
			s.T(), "UpdatePasswordWithCode", mock.Anything, mock.Anything, mock.Anything,
			mock.Anything,
		)
	})

	s.Run("should propagate the error when hashing the new password fails", func() {
		userID := fixtures.AuthCredsNewID()
		hashErr := errors.New("hash failure")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthCredsValidConfirmationCode(), nil)
		s.pw.On("Hash", fixtures.AuthCredsNewPassword).Return(nil, hashErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsNewPassword,
		)

		s.Require().ErrorIs(err, hashErr)
		s.gw.AssertNotCalled(s.T(), "UpdatePasswordWithCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run("should return a recovery code mismatch error when the code was consumed concurrently", func() {
		userID := fixtures.AuthCredsNewID()
		code := fixtures.AuthCredsValidConfirmationCode()
		hashed := fixtures.AuthCredsNewHash

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(code, nil)
		s.pw.On("Hash", fixtures.AuthCredsNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePasswordWithCode", mock.Anything, userID, code.ID, hashed).
			Return(false, nil)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsNewPassword,
		)

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
		s.gw.AssertExpectations(s.T())
		s.gw.AssertNotCalled(s.T(), "RevokeUserRefreshTokens", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when consuming the code and updating fails", func() {
		userID := fixtures.AuthCredsNewID()
		code := fixtures.AuthCredsValidConfirmationCode()
		hashed := fixtures.AuthCredsNewHash
		updateErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(code, nil)
		s.pw.On("Hash", fixtures.AuthCredsNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePasswordWithCode", mock.Anything, userID, code.ID, hashed).
			Return(false, updateErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsNewPassword,
		)

		s.Require().ErrorIs(err, updateErr)
		s.gw.AssertNotCalled(s.T(), "RevokeUserRefreshTokens", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when revoking the user refresh tokens fails", func() {
		userID := fixtures.AuthCredsNewID()
		code := fixtures.AuthCredsValidConfirmationCode()
		hashed := fixtures.AuthCredsNewHash
		revokeErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(code, nil)
		s.pw.On("Hash", fixtures.AuthCredsNewPassword).Return(&hashed, nil)
		s.gw.On("UpdatePasswordWithCode", mock.Anything, userID, code.ID, hashed).
			Return(true, nil)
		s.gw.On("RevokeUserRefreshTokens", mock.Anything, userID).Return(revokeErr)

		err := s.uc.Execute(
			context.Background(), userID, fixtures.AuthCredsCode, fixtures.AuthCredsNewPassword,
		)

		s.Require().ErrorIs(err, revokeErr)
	})
}
