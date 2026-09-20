package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

func activeUser() *user.User {
	return &user.User{
		ID:       uuid.New(),
		Email:    "jane@example.com",
		Password: "stored-hash",
		Active:   true,
	}
}

type SignInUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	pw *mocks.Password
	j  *mocks.JWTAuth
}

func (s *SignInUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.pw = new(mocks.Password)
	s.j = new(mocks.JWTAuth)
}

func TestSignInUseCase(t *testing.T) {
	suite.Run(t, new(SignInUseCaseTestSuite))
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_Success() {
	s.Run("should sign in and return an access and refresh token pair", func() {
		u := activeUser()
		access := "access-token"
		refresh := "refresh-token"
		expiresAt := time.Now().Add(24 * time.Hour)

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "secret123", "stored-hash").Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().NoError(err)
		s.Require().NotNil(tokens)
		s.Equal(access, tokens.AccessToken)
		s.Equal(refresh, tokens.RefreshToken)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
		s.j.AssertExpectations(s.T())
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_SaveRefreshTokenFails() {
	s.Run("should return an error when saving the refresh token fails", func() {
		u := activeUser()
		access := "access-token"
		refresh := "refresh-token"
		expiresAt := time.Now().Add(24 * time.Hour)

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "secret123", "stored-hash").Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(errors.New("db down"))

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().Error(err)
		s.Nil(tokens)
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_UnknownUser() {
	s.Run("should return an unauthorized error when the user does not exist", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "nobody@example.com").
			Return(nil, nil)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "nobody@example.com", "secret123")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1010_INVALID_CREDENTIALS, *appErr.BusinessCode())
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_WrongPassword() {
	s.Run("should return an unauthorized error when the password does not match", func() {
		u := activeUser()
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "wrong", "stored-hash").Return(errors.New("mismatch"))

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "wrong")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_InactiveAccount() {
	s.Run("should return a forbidden error when the account is not confirmed", func() {
		u := activeUser()
		u.Active = false

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "secret123", "stored-hash").Return(nil)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ForbiddenKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1009_ACCOUNT_NOT_CONFIRMED, *appErr.BusinessCode())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_GatewayError() {
	s.Run("should propagate the error when fetching the user fails", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(nil, gwErr)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().ErrorIs(err, gwErr)
		s.Nil(tokens)
		s.pw.AssertNotCalled(s.T(), "Verify", mock.Anything, mock.Anything)
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_IssueAccessTokenError() {
	s.Run("should propagate the error when issuing the access token fails", func() {
		issueErr := errors.New("signing failure")

		u := activeUser()
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "secret123", "stored-hash").Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_IssueRefreshTokenError() {
	s.Run("should propagate the error when issuing the refresh token fails", func() {
		issueErr := errors.New("signing failure")

		u := activeUser()
		access := "access-token"
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "secret123", "stored-hash").Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})
}

func (s *SignInUseCaseTestSuite) TestSignInUseCase_RefreshTokenMissingExpiry() {
	s.Run("should return a dependency error when the issued refresh token has no expiry", func() {
		u := activeUser()
		access := "access-token"
		refresh := "refresh-token"
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, "jane@example.com").Return(u, nil)
		s.pw.On("Verify", "secret123", "stored-hash").Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &refresh, ExpiresAt: nil}, nil)

		uc := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j)

		tokens, err := uc.Execute(context.Background(), "jane@example.com", "secret123")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.DependencyKind, appErr.Kind())
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})
}
