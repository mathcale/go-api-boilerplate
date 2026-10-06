package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	usermocks "github.com/mathcale/go-api-boilerplate/internal/modules/user/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/usecases"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type ValidateRecoveryCodeUseCaseTestSuite struct {
	suite.Suite
	gw *usermocks.UserGateway
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(usermocks.UserGateway)
}

func TestValidateRecoveryCodeUseCase(t *testing.T) {
	suite.Run(t, new(ValidateRecoveryCodeUseCaseTestSuite))
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) TestSuccess() {
	s.Run("should succeed when the recovery code is valid and not expired", func() {
		userID := uuid.New()
		code := &domain.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Minute)}

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "the-code", domain.PurposePasswordRecovery).
			Return(code, nil)

		uc := usecases.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "the-code")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) TestInvalidCode() {
	s.Run("should return a mismatch business error when the code does not match", func() {
		userID := uuid.New()

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "bad", domain.PurposePasswordRecovery).
			Return(nil, nil)

		uc := usecases.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "bad")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1002_RECOVERY_CODE_MISMATCH, *appErr.BusinessCode())
	})
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) TestExpiredCode() {
	s.Run("should return an expired business error when the code has expired", func() {
		userID := uuid.New()
		code := &domain.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "old", domain.PurposePasswordRecovery).
			Return(code, nil)

		uc := usecases.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "old")

		s.Require().Error(err)
		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.BE1007_CONFIRMATION_CODE_EXPIRED, *appErr.BusinessCode())
	})
}

func (s *ValidateRecoveryCodeUseCaseTestSuite) TestGatewayError() {
	s.Run("should propagate the error returned by the gateway", func() {
		userID := uuid.New()
		gwErr := errors.New("db down")

		s.gw.On("GetValidConfirmationCode", mock.Anything, userID, "code", domain.PurposePasswordRecovery).
			Return(nil, gwErr)

		uc := usecases.NewValidateRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), userID, "code")

		s.Require().ErrorIs(err, gwErr)
	})
}
