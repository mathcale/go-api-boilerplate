package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/gateway"
	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
)

type UserGatewayTestSuite struct {
	suite.Suite
	repo        *mocks.UserRepository
	emailClient *mocks.EmailClient
	gateway     gateway.User
}

func (s *UserGatewayTestSuite) SetupTest() {
	s.repo = new(mocks.UserRepository)
	s.emailClient = new(mocks.EmailClient)

	s.gateway = NewGateway(
		s.repo,
		s.emailClient,
		"http://confirm",
		"http://recover",
	)
}

func (s *UserGatewayTestSuite) cleanMocks() {
	s.repo.ExpectedCalls = nil
	s.repo.Calls = nil
	s.emailClient.ExpectedCalls = nil
	s.emailClient.Calls = nil
}

func TestUserGateway(t *testing.T) {
	suite.Run(t, new(UserGatewayTestSuite))
}

func (s *UserGatewayTestSuite) TestUserExistsIncludingInactive() {
	s.Run("should return true when user exists including inactive users", func() {
		defer s.cleanMocks()

		truthy := true
		s.repo.
			On("ExistsIncludingInactive", mock.Anything, fixtures.InfraUserEmail).
			Return(&truthy, nil)

		exists, err := s.gateway.UserExistsIncludingInactive(
			context.Background(),
			fixtures.InfraUserEmail,
		)

		s.Require().NoError(err)
		s.True(*exists)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSaveUser() {
	s.Run("should save a new user along with its confirmation code", func() {
		defer s.cleanMocks()

		s.repo.
			On(
				"Save",
				mock.Anything,
				mock.AnythingOfType("models.User"),
				mock.AnythingOfType("models.ConfirmationCode"),
			).
			Return(nil)

		u := fixtures.InfraDomainUser()
		code := fixtures.InfraDomainConfirmationCode(u.ID)

		err := s.gateway.SaveUser(context.Background(), u, code)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestGetUserByEmailIncludingInactive() {
	s.Run("should return the user when found by email including inactive users", func() {
		defer s.cleanMocks()

		m := fixtures.InfraModelUser()

		s.repo.On("GetIncludingInactive", mock.Anything, fixtures.InfraUserEmail).Return(&m, nil)

		got, err := s.gateway.GetUserByEmailIncludingInactive(
			context.Background(),
			fixtures.InfraUserEmail,
		)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(m.ID, got.ID)
	})

	s.Run("should return nil when no user is found by email", func() {
		defer s.cleanMocks()

		s.repo.
			On("GetIncludingInactive", mock.Anything, fixtures.InfraUnknownEmail).
			Return(nil, nil)

		got, err := s.gateway.GetUserByEmailIncludingInactive(
			context.Background(),
			fixtures.InfraUnknownEmail,
		)

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		s.repo.
			On("GetIncludingInactive", mock.Anything, fixtures.InfraUserEmail).
			Return(nil, errors.New("db down"))

		got, err := s.gateway.GetUserByEmailIncludingInactive(
			context.Background(),
			fixtures.InfraUserEmail,
		)

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestGetUserByID() {
	s.Run("should return the user when found by id", func() {
		defer s.cleanMocks()

		m := fixtures.InfraModelUser()
		userID := m.ID

		s.repo.On("GetByID", mock.Anything, userID).Return(&m, nil)

		got, err := s.gateway.GetUserByID(context.Background(), userID)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(userID, got.ID)
	})

	s.Run("should return nil when no user is found by id", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("GetByID", mock.Anything, userID).Return(nil, nil)

		got, err := s.gateway.GetUserByID(context.Background(), userID)

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("db down"))

		got, err := s.gateway.GetUserByID(context.Background(), userID)

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestGetUserAuthState() {
	s.Run("should return the auth state of the user", func() {
		defer s.cleanMocks()

		m := fixtures.InfraModelUserAuthState()
		userID := uuid.New()
		s.repo.On("GetAuthState", mock.Anything, userID).Return(&m, nil)

		got, err := s.gateway.GetUserAuthState(context.Background(), userID)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(m.TokenVersion, got.TokenVersion)
		s.Equal(m.Active, got.Active)
	})

	s.Run("should return nil when the user does not exist", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("GetAuthState", mock.Anything, userID).Return(nil, nil)

		got, err := s.gateway.GetUserAuthState(context.Background(), userID)

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("GetAuthState", mock.Anything, userID).Return(nil, errors.New("db down"))

		got, err := s.gateway.GetUserAuthState(context.Background(), userID)

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestUpdatePasswordWithCode() {
	s.Run("should update the password consuming the code", func() {
		defer s.cleanMocks()

		userID, codeID := uuid.New(), uuid.New()
		s.repo.On("UpdatePasswordWithCode", mock.Anything, userID, codeID, "new-hash").
			Return(true, nil)

		updated, err := s.gateway.UpdatePasswordWithCode(
			context.Background(), userID, codeID, "new-hash",
		)

		s.Require().NoError(err)
		s.True(updated)
		s.repo.AssertExpectations(s.T())
	})

	s.Run("should report not updated when the code was already consumed", func() {
		defer s.cleanMocks()

		userID, codeID := uuid.New(), uuid.New()
		s.repo.On("UpdatePasswordWithCode", mock.Anything, userID, codeID, "new-hash").
			Return(false, nil)

		updated, err := s.gateway.UpdatePasswordWithCode(
			context.Background(), userID, codeID, "new-hash",
		)

		s.Require().NoError(err)
		s.False(updated)
	})

	s.Run("should return the repository error", func() {
		defer s.cleanMocks()

		userID, codeID := uuid.New(), uuid.New()
		s.repo.On("UpdatePasswordWithCode", mock.Anything, userID, codeID, "new-hash").
			Return(false, errors.New("db down"))

		updated, err := s.gateway.UpdatePasswordWithCode(
			context.Background(), userID, codeID, "new-hash",
		)

		s.Require().Error(err)
		s.False(updated)
	})
}

func (s *UserGatewayTestSuite) TestActivateUserWithCode() {
	s.Run("should activate the user consuming the code", func() {
		defer s.cleanMocks()

		userID, codeID := uuid.New(), uuid.New()
		s.repo.On("ActivateUserWithCode", mock.Anything, userID, codeID).Return(true, nil)

		activated, err := s.gateway.ActivateUserWithCode(context.Background(), userID, codeID)

		s.Require().NoError(err)
		s.True(activated)
		s.repo.AssertExpectations(s.T())
	})

	s.Run("should report not activated when the code was already consumed", func() {
		defer s.cleanMocks()

		userID, codeID := uuid.New(), uuid.New()
		s.repo.On("ActivateUserWithCode", mock.Anything, userID, codeID).Return(false, nil)

		activated, err := s.gateway.ActivateUserWithCode(context.Background(), userID, codeID)

		s.Require().NoError(err)
		s.False(activated)
	})

	s.Run("should return the repository error", func() {
		defer s.cleanMocks()

		userID, codeID := uuid.New(), uuid.New()
		s.repo.On("ActivateUserWithCode", mock.Anything, userID, codeID).
			Return(false, errors.New("db down"))

		activated, err := s.gateway.ActivateUserWithCode(context.Background(), userID, codeID)

		s.Require().Error(err)
		s.False(activated)
	})
}

func (s *UserGatewayTestSuite) TestSaveConfirmationCode() {
	s.Run("should save a confirmation code", func() {
		defer s.cleanMocks()

		s.repo.
			On(
				"SaveConfirmationCode",
				mock.Anything,
				mock.AnythingOfType("models.ConfirmationCode"),
			).
			Return(nil)

		code := fixtures.InfraDomainRecoveryCode(uuid.New())

		err := s.gateway.SaveConfirmationCode(context.Background(), code)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestGetValidConfirmationCode() {
	s.Run("should return the confirmation code when found and valid", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		m := fixtures.InfraModelConfirmationCode(userID)

		s.repo.
			On("GetConfirmationCode", mock.Anything, userID, "the-code", "account_confirmation").
			Return(&m, nil)

		got, err := s.gateway.GetValidConfirmationCode(
			context.Background(),
			userID,
			"the-code",
			user.PurposeAccountConfirmation,
		)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(m.ID, got.ID)
	})

	s.Run("should return nil when no valid confirmation code is found", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.
			On("GetConfirmationCode", mock.Anything, userID, "nope", "account_confirmation").
			Return(nil, nil)

		got, err := s.gateway.GetValidConfirmationCode(
			context.Background(),
			userID,
			"nope",
			user.PurposeAccountConfirmation,
		)

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.
			On("GetConfirmationCode", mock.Anything, userID, "the-code", "account_confirmation").
			Return(nil, errors.New("db down"))

		got, err := s.gateway.GetValidConfirmationCode(
			context.Background(),
			userID,
			"the-code",
			user.PurposeAccountConfirmation,
		)

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestSendConfirmationEmail() {
	s.Run("should send the account confirmation email", func() {
		defer s.cleanMocks()

		s.emailClient.On("Send", mock.AnythingOfType("email.SendInput")).Return(nil)

		u := fixtures.InfraDomainUser()
		code := fixtures.InfraDomainConfirmationCode(u.ID)

		err := s.gateway.SendConfirmationEmail(context.Background(), u, code)

		s.Require().NoError(err)
		s.emailClient.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSendRecoveryEmail() {
	s.Run("should send the password recovery email", func() {
		defer s.cleanMocks()

		s.emailClient.On("Send", mock.AnythingOfType("email.SendInput")).Return(nil)

		u := fixtures.InfraDomainUser()
		code := fixtures.InfraDomainRecoveryCode(u.ID)

		err := s.gateway.SendRecoveryEmail(context.Background(), u, code)

		s.Require().NoError(err)
		s.emailClient.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSaveRefreshToken() {
	s.Run("should save a new refresh token", func() {
		defer s.cleanMocks()

		s.repo.
			On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("models.RefreshToken")).
			Return(nil)

		rt := fixtures.InfraDomainRefreshToken()

		err := s.gateway.SaveRefreshToken(context.Background(), rt)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestGetRefreshToken() {
	s.Run("should return the refresh token when found", func() {
		defer s.cleanMocks()

		m := fixtures.InfraModelRefreshToken()
		tokenID := m.ID

		s.repo.On("GetRefreshToken", mock.Anything, tokenID).Return(&m, nil)

		got, err := s.gateway.GetRefreshToken(context.Background(), tokenID)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(tokenID, got.ID)
	})

	s.Run("should return nil when no refresh token is found", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()
		s.repo.On("GetRefreshToken", mock.Anything, tokenID).Return(nil, nil)

		got, err := s.gateway.GetRefreshToken(context.Background(), tokenID)

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()
		s.repo.On("GetRefreshToken", mock.Anything, tokenID).Return(nil, errors.New("db down"))

		got, err := s.gateway.GetRefreshToken(context.Background(), tokenID)

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestMarkRefreshTokenUsed() {
	s.Run("should report the refresh token as claimed", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()
		s.repo.On("MarkRefreshTokenUsed", mock.Anything, tokenID).Return(true, nil)

		claimed, err := s.gateway.MarkRefreshTokenUsed(context.Background(), tokenID)

		s.Require().NoError(err)
		s.True(claimed)
		s.repo.AssertExpectations(s.T())
	})

	s.Run("should report the refresh token as not claimed when it was lost", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()
		s.repo.On("MarkRefreshTokenUsed", mock.Anything, tokenID).Return(false, nil)

		claimed, err := s.gateway.MarkRefreshTokenUsed(context.Background(), tokenID)

		s.Require().NoError(err)
		s.False(claimed)
	})

	s.Run("should return the repository error", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()
		s.repo.On("MarkRefreshTokenUsed", mock.Anything, tokenID).
			Return(false, errors.New("db down"))

		claimed, err := s.gateway.MarkRefreshTokenUsed(context.Background(), tokenID)

		s.Require().Error(err)
		s.False(claimed)
	})
}

func (s *UserGatewayTestSuite) TestRevokeUserRefreshTokens() {
	s.Run("should revoke every refresh token of the user", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("RevokeUserRefreshTokens", mock.Anything, userID).Return(nil)

		err := s.gateway.RevokeUserRefreshTokens(context.Background(), userID)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestRevokeRefreshTokenFamily() {
	s.Run("should revoke the entire refresh token family", func() {
		defer s.cleanMocks()

		familyID := uuid.New()
		s.repo.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(nil)

		err := s.gateway.RevokeRefreshTokenFamily(context.Background(), familyID)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}
