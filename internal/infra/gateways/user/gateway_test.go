package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/gateway"
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
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
	s.gateway = NewGateway(s.repo, s.emailClient, "http://confirm", "http://recover")
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
		s.repo.On("ExistsIncludingInactive", mock.Anything, "jane@example.com").Return(&truthy, nil)

		exists, err := s.gateway.UserExistsIncludingInactive(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.True(*exists)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSaveUser() {
	s.Run("should save a new user along with its confirmation code", func() {
		defer s.cleanMocks()

		s.repo.
			On("Save", mock.Anything, mock.AnythingOfType("models.User"), mock.AnythingOfType("models.ConfirmationCode")).
			Return(nil)

		u := user.User{
			ID:    uuid.New(),
			Email: "jane@example.com",
		}

		code := user.NewConfirmationCode(u.ID, user.PurposeAccountConfirmation, user.AccountConfirmationTTL)

		err := s.gateway.SaveUser(context.Background(), u, code)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestGetUserByEmailIncludingInactive() {
	s.Run("should return the user when found by email including inactive users", func() {
		defer s.cleanMocks()

		m := &models.User{
			ID:    uuid.New(),
			Email: "jane@example.com",
		}

		s.repo.On("GetIncludingInactive", mock.Anything, "jane@example.com").Return(m, nil)

		got, err := s.gateway.GetUserByEmailIncludingInactive(context.Background(), "jane@example.com")

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(m.ID, got.ID)
	})

	s.Run("should return nil when no user is found by email", func() {
		defer s.cleanMocks()

		s.repo.On("GetIncludingInactive", mock.Anything, "nobody@example.com").Return(nil, nil)

		got, err := s.gateway.GetUserByEmailIncludingInactive(context.Background(), "nobody@example.com")

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		s.repo.
			On("GetIncludingInactive", mock.Anything, "jane@example.com").
			Return(nil, errors.New("db down"))

		got, err := s.gateway.GetUserByEmailIncludingInactive(context.Background(), "jane@example.com")

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestGetUserByID() {
	s.Run("should return the user when found by id", func() {
		defer s.cleanMocks()

		userID := uuid.New()

		m := &models.User{
			ID: userID,
		}

		s.repo.On("GetByID", mock.Anything, userID).Return(m, nil)

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

func (s *UserGatewayTestSuite) TestUpdatePassword() {
	s.Run("should update the user's password", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("UpdatePassword", mock.Anything, userID, "new-hash").Return(nil)

		err := s.gateway.UpdatePassword(context.Background(), userID, "new-hash")

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestActivateUser() {
	s.Run("should activate the user", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("ActivateUser", mock.Anything, userID).Return(nil)

		err := s.gateway.ActivateUser(context.Background(), userID)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSaveConfirmationCode() {
	s.Run("should save a confirmation code", func() {
		defer s.cleanMocks()

		s.repo.
			On("SaveConfirmationCode", mock.Anything, mock.AnythingOfType("models.ConfirmationCode")).
			Return(nil)

		code := user.NewConfirmationCode(uuid.New(), user.PurposePasswordRecovery, user.PasswordRecoveryTTL)

		err := s.gateway.SaveConfirmationCode(context.Background(), code)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestGetValidConfirmationCode() {
	s.Run("should return the confirmation code when found and valid", func() {
		defer s.cleanMocks()

		userID := uuid.New()

		m := &models.ConfirmationCode{
			ID:      uuid.New(),
			UserID:  userID,
			Purpose: "account_confirmation",
		}

		s.repo.
			On("GetConfirmationCode", mock.Anything, userID, "the-code", "account_confirmation").
			Return(m, nil)

		got, err := s.gateway.GetValidConfirmationCode(context.Background(), userID, "the-code", user.PurposeAccountConfirmation)

		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Equal(m.ID, got.ID)
	})

	s.Run("should return nil when no valid confirmation code is found", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("GetConfirmationCode", mock.Anything, userID, "nope", "account_confirmation").
			Return(nil, nil)

		got, err := s.gateway.GetValidConfirmationCode(context.Background(), userID, "nope", user.PurposeAccountConfirmation)

		s.Require().NoError(err)
		s.Nil(got)
	})

	s.Run("should return an error when the repository fails", func() {
		defer s.cleanMocks()

		userID := uuid.New()
		s.repo.On("GetConfirmationCode", mock.Anything, userID, "the-code", "account_confirmation").
			Return(nil, errors.New("db down"))

		got, err := s.gateway.GetValidConfirmationCode(context.Background(), userID, "the-code", user.PurposeAccountConfirmation)

		s.Require().Error(err)
		s.Nil(got)
	})
}

func (s *UserGatewayTestSuite) TestMarkConfirmationCodeUsed() {
	s.Run("should mark the confirmation code as used", func() {
		defer s.cleanMocks()

		codeID := uuid.New()
		s.repo.On("MarkConfirmationCodeUsed", mock.Anything, codeID).Return(nil)

		err := s.gateway.MarkConfirmationCodeUsed(context.Background(), codeID)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSendConfirmationEmail() {
	s.Run("should send the account confirmation email", func() {
		defer s.cleanMocks()

		s.emailClient.On("Send", mock.AnythingOfType("email.SendInput")).Return(nil)

		u := user.User{
			ID:    uuid.New(),
			Name:  "Jane",
			Email: "jane@example.com",
		}

		code := user.NewConfirmationCode(u.ID, user.PurposeAccountConfirmation, user.AccountConfirmationTTL)

		err := s.gateway.SendConfirmationEmail(context.Background(), u, code)

		s.Require().NoError(err)
		s.emailClient.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSendRecoveryEmail() {
	s.Run("should send the password recovery email", func() {
		defer s.cleanMocks()

		s.emailClient.On("Send", mock.AnythingOfType("email.SendInput")).Return(nil)

		u := user.User{
			ID:    uuid.New(),
			Name:  "Jane",
			Email: "jane@example.com",
		}

		code := user.NewConfirmationCode(u.ID, user.PurposePasswordRecovery, user.PasswordRecoveryTTL)

		err := s.gateway.SendRecoveryEmail(context.Background(), u, code)

		s.Require().NoError(err)
		s.emailClient.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestSaveRefreshToken() {
	s.Run("should save a new refresh token", func() {
		defer s.cleanMocks()

		s.repo.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("models.RefreshToken")).Return(nil)

		rt := user.NewRefreshTokenFamily(uuid.New(), uuid.New(), time.Now().Add(time.Hour))

		err := s.gateway.SaveRefreshToken(context.Background(), rt)

		s.Require().NoError(err)
		s.repo.AssertExpectations(s.T())
	})
}

func (s *UserGatewayTestSuite) TestGetRefreshToken() {
	s.Run("should return the refresh token when found", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()

		m := &models.RefreshToken{
			ID: tokenID,
		}

		s.repo.On("GetRefreshToken", mock.Anything, tokenID).Return(m, nil)

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
	s.Run("should mark the refresh token as used", func() {
		defer s.cleanMocks()

		tokenID := uuid.New()
		s.repo.On("MarkRefreshTokenUsed", mock.Anything, tokenID).Return(nil)

		err := s.gateway.MarkRefreshTokenUsed(context.Background(), tokenID)

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
