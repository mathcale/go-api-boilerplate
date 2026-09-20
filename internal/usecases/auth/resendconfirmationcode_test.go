package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type ResendConfirmationCodeUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
}

func (s *ResendConfirmationCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
}

func TestResendConfirmationCodeUseCase(t *testing.T) {
	suite.Run(t, new(ResendConfirmationCodeUseCaseTestSuite))
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestSuccess() {
	s.Run("should generate and send a new confirmation code to an inactive user", func() {
		target := &user.User{ID: uuid.New(), Email: "jane@example.com", Active: false}

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("user.ConfirmationCode")).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, *target, mock.AnythingOfType("user.ConfirmationCode")).
			Return(nil)

		uc := auth.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestUnknownUser() {
	s.Run("should silently succeed without sending a code when the email is unknown", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "nobody@example.com").Return(nil, nil)

		uc := auth.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), "nobody@example.com")

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestAlreadyActive() {
	s.Run("should silently succeed without sending a code when the user is already active", func() {
		target := &user.User{ID: uuid.New(), Email: "jane@example.com", Active: true}

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)

		uc := auth.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestGatewayError() {
	s.Run("should propagate the error returned when looking up the user", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(nil, gwErr)

		uc := auth.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().ErrorIs(err, gwErr)
	})
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestSaveConfirmationCodeError() {
	s.Run("should propagate the error and skip sending the email when saving the code fails", func() {
		target := &user.User{ID: uuid.New(), Email: "jane@example.com", Active: false}
		saveErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("user.ConfirmationCode")).
			Return(saveErr)

		uc := auth.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().ErrorIs(err, saveErr)
		s.gw.AssertNotCalled(s.T(), "SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything)
	})
}

// Email delivery failures must not fail the flow: the user can request another code.
func (s *ResendConfirmationCodeUseCaseTestSuite) TestEmailFailureIsNonFatal() {
	s.Run("should log the error but still succeed when sending the confirmation email fails", func() {
		target := &user.User{ID: uuid.New(), Email: "jane@example.com", Active: false}
		log := new(mocks.Logger)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(target, nil)
		s.gw.On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("user.ConfirmationCode")).Return(nil)
		s.gw.On("SendConfirmationEmail", mock.Anything, *target, mock.AnythingOfType("user.ConfirmationCode")).
			Return(errors.New("smtp down"))

		uc := auth.NewResendConfirmationCodeUseCase(log, s.gw)

		err := uc.Execute(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		log.AssertExpectations(s.T())
	})
}
