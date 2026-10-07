package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	usermocks "github.com/mathcale/go-api-boilerplate/internal/modules/user/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/usecases"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type SetRecoveryCodeUseCaseTestSuite struct {
	suite.Suite
	gw *usermocks.UserGateway
}

func (s *SetRecoveryCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(usermocks.UserGateway)
}

func TestSetRecoveryCodeUseCase(t *testing.T) {
	suite.Run(t, new(SetRecoveryCodeUseCaseTestSuite))
}

func (s *SetRecoveryCodeUseCaseTestSuite) TestSuccess() {
	s.Run("should generate and send a recovery code to a known user", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com"}

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("domain.ConfirmationCode")).Return(nil)
		s.gw.On("SendRecoveryEmail", mock.Anything, *target, mock.AnythingOfType("domain.ConfirmationCode")).
			Return(nil)

		uc := usecases.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw, testRecoveryTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *SetRecoveryCodeUseCaseTestSuite) TestUnknownUser() {
	s.Run("should silently succeed without sending a code when the email is unknown", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "nobody@example.com").Return(nil, nil)

		uc := usecases.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw, testRecoveryTTL)

		err := uc.Execute(context.Background(), "nobody@example.com")

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})
}

func (s *SetRecoveryCodeUseCaseTestSuite) TestGatewayError() {
	s.Run("should propagate the error returned when looking up the user", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(nil, gwErr)

		uc := usecases.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw, testRecoveryTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().ErrorIs(err, gwErr)
	})
}

func (s *SetRecoveryCodeUseCaseTestSuite) TestSaveConfirmationCodeError() {
	s.Run("should propagate the error and skip sending the email when saving the code fails", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com"}
		saveErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("domain.ConfirmationCode")).
			Return(saveErr)

		uc := usecases.NewSetRecoveryCodeUseCase(mocks.NoopLogger{}, s.gw, testRecoveryTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().ErrorIs(err, saveErr)
		s.gw.AssertNotCalled(s.T(), "SendRecoveryEmail", mock.Anything, mock.Anything, mock.Anything)
	})
}

// Email delivery failures must not fail the flow: the user can request recovery again.
func (s *SetRecoveryCodeUseCaseTestSuite) TestEmailFailureIsNonFatal() {
	s.Run("should log the error but still succeed when sending the recovery email fails", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com"}
		log := new(mocks.Logger)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("domain.ConfirmationCode")).Return(nil)
		s.gw.On("SendRecoveryEmail", mock.Anything, *target, mock.AnythingOfType("domain.ConfirmationCode")).
			Return(errors.New("smtp down"))

		uc := usecases.NewSetRecoveryCodeUseCase(log, s.gw, testRecoveryTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		log.AssertExpectations(s.T())
	})
}
