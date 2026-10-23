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

type ValidateRecoveryCodeUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestValidateRecoveryCodeUseCase(t *testing.T) {
	suite.Run(t, new(ValidateRecoveryCodeUseCaseTestSuite))
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) TestExecute() {
	s.Run("should succeed when the recovery code is valid and not expired", func() {
		userID := fixtures.AuthCredsNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "the-code",
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthCredsValidConfirmationCode(), nil)

		uc := auth.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should return a mismatch business error when the code does not match", func() {
		userID := fixtures.AuthCredsNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "bad",
			user.PurposePasswordRecovery,
		).Return(nil, nil)

		uc := auth.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "bad")

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
	})

	s.Run("should return an expired business error when the code has expired", func() {
		userID := fixtures.AuthCredsNewID()

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, "old",
			user.PurposePasswordRecovery,
		).Return(fixtures.AuthCredsExpiredConfirmationCode(), nil)

		uc := auth.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "old")

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
	})

	s.Run("should propagate the error returned by the gateway", func() {
		userID := fixtures.AuthCredsNewID()
		gwErr := errors.New("db down")

		s.gw.On(
			"GetValidConfirmationCode", mock.Anything, userID, fixtures.AuthCredsCode,
			user.PurposePasswordRecovery,
		).Return(nil, gwErr)

		uc := auth.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, fixtures.AuthCredsCode)

		s.Require().ErrorIs(err, gwErr)
	})
}
