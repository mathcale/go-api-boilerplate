package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
	"go-api-boilerplate/internal/usecases/auth"
)

type SetRecoveryCodeUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
}

func (s *SetRecoveryCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
}

func (s *SetRecoveryCodeUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestSetRecoveryCodeUseCase(t *testing.T) {
	suite.Run(t, new(SetRecoveryCodeUseCaseTestSuite))
}

func (s *SetRecoveryCodeUseCaseTestSuite) TestExecute() {
	s.Run("should generate and send a recovery code to a known user", func() {
		target := fixtures.AuthCredsRecoveryUser()

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(target, nil)
		s.gw.On(
			"SaveConfirmationCode", mock.Anything, mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On(
			"SendRecoveryEmail", mock.Anything, *target,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)

		uc := auth.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), fixtures.AuthCredsEmail)

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should normalise the email before looking up the user", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(nil, nil)

		uc := auth.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), fixtures.AuthCredsMixedCaseEmail)

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should silently succeed without sending a code when the email is unknown", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsUnknownEmail).
			Return(nil, nil)

		uc := auth.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), fixtures.AuthCredsUnknownEmail)

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})

	s.Run("should log the error but still succeed when sending the recovery email fails", func() {
		target := fixtures.AuthCredsRecoveryUser()
		log := new(mocks.Logger)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(target, nil)
		s.gw.On(
			"SaveConfirmationCode", mock.Anything, mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On(
			"SendRecoveryEmail", mock.Anything, *target,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(errors.New("smtp down"))

		uc := auth.NewSetRecoveryCodeUseCase(log, s.gw)

		err := uc.Execute(context.Background(), fixtures.AuthCredsEmail)

		s.Require().NoError(err)
		log.AssertExpectations(s.T())
	})

	s.Run("should propagate the error returned when looking up the user", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(nil, gwErr)

		uc := auth.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), fixtures.AuthCredsEmail)

		s.Require().ErrorIs(err, gwErr)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error and skip the email when saving the code fails", func() {
		target := fixtures.AuthCredsRecoveryUser()
		saveErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(target, nil)
		s.gw.On(
			"SaveConfirmationCode", mock.Anything, mock.AnythingOfType("user.ConfirmationCode"),
		).Return(saveErr)

		uc := auth.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), fixtures.AuthCredsEmail)

		s.Require().ErrorIs(err, saveErr)
		s.gw.AssertNotCalled(
			s.T(), "SendRecoveryEmail", mock.Anything, mock.Anything, mock.Anything,
		)
	})
}
