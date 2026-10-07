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

type ResendConfirmationCodeUseCaseTestSuite struct {
	suite.Suite
	gw *usermocks.UserGateway
}

func (s *ResendConfirmationCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(usermocks.UserGateway)
}

func TestResendConfirmationCodeUseCase(t *testing.T) {
	suite.Run(t, new(ResendConfirmationCodeUseCaseTestSuite))
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestSuccess() {
	s.Run("should generate and send a new confirmation code to an inactive user", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com", Active: false}

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("domain.ConfirmationCode")).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, *target, mock.AnythingOfType("domain.ConfirmationCode")).
			Return(nil)

		uc := usecases.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw, testConfirmationTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestUnknownUser() {
	s.Run("should silently succeed without sending a code when the email is unknown", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "nobody@example.com").Return(nil, nil)

		uc := usecases.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw, testConfirmationTTL)

		err := uc.Execute(context.Background(), "nobody@example.com")

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestAlreadyActive() {
	s.Run("should silently succeed without sending a code when the user is already active", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com", Active: true}

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)

		uc := usecases.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw, testConfirmationTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestGatewayError() {
	s.Run("should propagate the error returned when looking up the user", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(nil, gwErr)

		uc := usecases.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw, testConfirmationTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().ErrorIs(err, gwErr)
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestSaveConfirmationCodeError() {
	s.Run("should propagate the error and skip sending the email when saving the code fails", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com", Active: false}
		saveErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("domain.ConfirmationCode")).
			Return(saveErr)

		uc := usecases.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw, testConfirmationTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().ErrorIs(err, saveErr)
		s.gw.AssertNotCalled(s.T(), "SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything)
	})
}

// Email delivery failures must not fail the flow: the user can request another code.
func (s *ResendConfirmationCodeUseCaseTestSuite) TestEmailFailureIsNonFatal() {
	s.Run("should log the error but still succeed when sending the confirmation email fails", func() {
		target := &domain.User{ID: uuid.New(), Email: "jane@example.com", Active: false}
		log := new(mocks.Logger)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("domain.ConfirmationCode")).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, *target, mock.AnythingOfType("domain.ConfirmationCode")).
			Return(errors.New("smtp down"))

		uc := usecases.NewResendConfirmationCodeUseCase(log, s.gw, testConfirmationTTL)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		log.AssertExpectations(s.T())
	})
}
