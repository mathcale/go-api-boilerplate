package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
	"go-api-boilerplate/internal/usecases/auth"
)

type ResendConfirmationCodeUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	uc auth.ResendConfirmationCodeUseCase
}

func (s *ResendConfirmationCodeUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.uc = auth.NewResendConfirmationCodeUseCase(mocks.NoopLogger{}, s.gw)
}

func (s *ResendConfirmationCodeUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestResendConfirmationCodeUseCase(t *testing.T) {
	suite.Run(t, new(ResendConfirmationCodeUseCaseTestSuite))
}

func (s *ResendConfirmationCodeUseCaseTestSuite) TestExecute() {
	s.Run("should generate and send a new confirmation code to an inactive user", func() {
		target := fixtures.AuthAUser(uuid.New(), false)

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthAEmail).
			Return(target, nil)
		s.gw.On(
			"SaveConfirmationCode", mock.Anything,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On(
			"SendConfirmationEmail", mock.Anything, *target,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)

		err := s.uc.Execute(context.Background(), fixtures.AuthAEmail)

		s.Require().NoError(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should normalise the email before looking up the user", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthAEmail).
			Return(nil, nil)

		_ = s.uc.Execute(context.Background(), " Jane@Example.COM ")

		s.gw.AssertExpectations(s.T())
	})

	s.Run("should log the error but still succeed when sending the email fails", func() {
		target := fixtures.AuthAUser(uuid.New(), false)
		log := new(mocks.Logger)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
		uc := auth.NewResendConfirmationCodeUseCase(log, s.gw)

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthAEmail).
			Return(target, nil)
		s.gw.On(
			"SaveConfirmationCode", mock.Anything,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(nil)
		s.gw.On(
			"SendConfirmationEmail", mock.Anything, *target,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(errors.New("smtp down"))

		err := uc.Execute(context.Background(), fixtures.AuthAEmail)

		s.Require().NoError(err)
		log.AssertExpectations(s.T())
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should silently succeed without sending a code when the email is unknown", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "nobody@example.com").
			Return(nil, nil)

		err := s.uc.Execute(context.Background(), "nobody@example.com")

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})

	s.Run("should silently succeed without sending a code when the user is active", func() {
		target := fixtures.AuthAUser(uuid.New(), true)

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthAEmail).
			Return(target, nil)

		err := s.uc.Execute(context.Background(), fixtures.AuthAEmail)

		s.Require().NoError(err)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error returned when looking up the user", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthAEmail).
			Return(nil, gwErr)

		err := s.uc.Execute(context.Background(), fixtures.AuthAEmail)

		s.Require().ErrorIs(err, gwErr)
		s.gw.AssertNotCalled(s.T(), "SaveConfirmationCode", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error and skip the email when saving the code fails", func() {
		target := fixtures.AuthAUser(uuid.New(), false)
		saveErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthAEmail).
			Return(target, nil)
		s.gw.On(
			"SaveConfirmationCode", mock.Anything,
			mock.AnythingOfType("user.ConfirmationCode"),
		).Return(saveErr)

		err := s.uc.Execute(context.Background(), fixtures.AuthAEmail)

		s.Require().ErrorIs(err, saveErr)
		s.gw.AssertNotCalled(
			s.T(), "SendConfirmationEmail", mock.Anything, mock.Anything, mock.Anything,
		)
	})
}
