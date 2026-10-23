package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/pkg/jwt"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
	"go-api-boilerplate/internal/usecases/auth"
)

type SignInUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	pw *mocks.Password
	j  *mocks.JWTAuth
	l  *mocks.AttemptLimiter
}

func (s *SignInUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.pw = new(mocks.Password)
	s.j = new(mocks.JWTAuth)
	s.l = new(mocks.AttemptLimiter)
}

func (s *SignInUseCaseTestSuite) newUseCase() auth.SignInUseCase {
	s.T().Helper()

	dummy := fixtures.AuthCredsDummyHash
	s.pw.On("Hash", mock.AnythingOfType("string")).Return(&dummy, nil).Once()

	if len(s.l.ExpectedCalls) == 0 {
		s.l.On("Allow", mock.Anything).Return(true)
	}

	uc, err := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j, s.l)
	s.Require().NoError(err)

	return uc
}

func (s *SignInUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestSignInUseCase(t *testing.T) {
	suite.Run(t, new(SignInUseCaseTestSuite))
}

func (s *SignInUseCaseTestSuite) TestNewSignInUseCase() {
	s.Run("should hash a dummy password once at construction", func() {
		s.newUseCase()

		s.pw.AssertNumberOfCalls(s.T(), "Hash", 1)
	})

	s.Run("should return the error when the dummy hash cannot be generated", func() {
		hashErr := errors.New("hash failure")
		s.pw.On("Hash", mock.AnythingOfType("string")).Return(nil, hashErr)

		uc, err := auth.NewSignInUseCase(mocks.NoopLogger{}, s.gw, s.pw, s.j, s.l)

		s.Nil(uc)
		s.ErrorIs(err, hashErr)
	})
}

func (s *SignInUseCaseTestSuite) TestExecute() {
	s.Run("should issue both tokens carrying the user's current token version", func() {
		hasVersion := mock.MatchedBy(func(p jwt.IssueTokenParams) bool {
			return p.TokenVersion == 6
		})

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUserWithVersion(6), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)
		s.j.On("IssueAccessToken", hasVersion).Return(fixtures.AuthCredsAccessJWT(), nil)
		s.j.On("IssueRefreshToken", hasVersion).Return(fixtures.AuthCredsRefreshJWT(), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		uc := s.newUseCase()

		_, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().NoError(err)
		s.j.AssertExpectations(s.T())
	})

	s.Run("should normalise the email before looking up the user", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(nil, nil)
		s.pw.On("Verify", mock.Anything, mock.Anything).Return(errors.New("mismatch"))

		uc := s.newUseCase()

		_, err := uc.Execute(
			context.Background(), fixtures.AuthCredsMixedCaseEmail, fixtures.AuthCredsPassword,
		)

		s.Require().Error(err)
		s.gw.AssertExpectations(s.T())
	})

	s.Run("should sign in and return an access and refresh token pair", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsAccessJWT(), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsRefreshJWT(), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().NoError(err)
		s.Require().NotNil(tokens)
		s.Equal(fixtures.AuthCredsAccessToken, tokens.AccessToken)
		s.Equal(fixtures.AuthCredsRefreshToken, tokens.RefreshToken)
		s.gw.AssertExpectations(s.T())
		s.pw.AssertExpectations(s.T())
		s.j.AssertExpectations(s.T())
	})

	s.Run("should check the limiter with the normalised email", func() {
		s.l.On("Allow", fixtures.AuthCredsEmail).Return(true)
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(nil, nil)
		s.pw.On("Verify", mock.Anything, mock.Anything).Return(errors.New("mismatch"))

		uc := s.newUseCase()

		_, err := uc.Execute(
			context.Background(), fixtures.AuthCredsMixedCaseEmail, fixtures.AuthCredsPassword,
		)

		s.Require().Error(err)
		s.l.AssertCalled(s.T(), "Allow", fixtures.AuthCredsEmail)
	})

	s.Run("should return a too-many-requests error when the limiter rejects", func() {
		s.l.On("Allow", fixtures.AuthCredsEmail).Return(false)

		uc := s.newUseCase()

		tokens, err := uc.Execute(
			context.Background(), fixtures.AuthCredsMixedCaseEmail, fixtures.AuthCredsPassword,
		)

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.TooManyRequests, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0005_RATE_LIMITED, *appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "GetUserByEmailIncludingInactive", mock.Anything, mock.Anything)
		s.pw.AssertNotCalled(s.T(), "Verify", mock.Anything, mock.Anything)
	})

	s.Run("should return an unauthorized error when the user does not exist", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsUnknownEmail).
			Return(nil, nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsDummyHash).
			Return(errors.New("mismatch"))

		uc := s.newUseCase()

		tokens, err := uc.Execute(
			context.Background(), fixtures.AuthCredsUnknownEmail, fixtures.AuthCredsPassword,
		)

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1010_INVALID_CREDENTIALS, *appErr.BusinessCode())
		s.pw.AssertCalled(s.T(), "Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsDummyHash)
	})

	s.Run("should return an unauthorized error when the password does not match", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsWrongPassword, fixtures.AuthCredsStoredHash).
			Return(errors.New("mismatch"))

		uc := s.newUseCase()

		tokens, err := uc.Execute(
			context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsWrongPassword,
		)

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1010_INVALID_CREDENTIALS, *appErr.BusinessCode())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should return a forbidden error when the account is not confirmed", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsInactiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ForbiddenKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE1009_ACCOUNT_NOT_CONFIRMED, *appErr.BusinessCode())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should propagate the error when fetching the user fails", func() {
		gwErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(nil, gwErr)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().ErrorIs(err, gwErr)
		s.Nil(tokens)
		s.pw.AssertNotCalled(s.T(), "Verify", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when issuing the access token fails", func() {
		issueErr := errors.New("signing failure")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})

	s.Run("should propagate the error when issuing the refresh token fails", func() {
		issueErr := errors.New("signing failure")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsAccessJWT(), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should return a dependency error when the refresh token has no expiry", func() {
		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsAccessJWT(), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsRefreshJWTWithoutExpiry(), nil)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.DependencyKind, appErr.Kind())
		s.Nil(appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should return an error when saving the refresh token fails", func() {
		saveErr := errors.New("db down")

		s.gw.On("GetUserByEmailIncludingInactive", mock.Anything, fixtures.AuthCredsEmail).
			Return(fixtures.AuthCredsActiveUser(), nil)
		s.pw.On("Verify", fixtures.AuthCredsPassword, fixtures.AuthCredsStoredHash).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsAccessJWT(), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthCredsRefreshJWT(), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(saveErr)

		uc := s.newUseCase()

		tokens, err := uc.Execute(context.Background(), fixtures.AuthCredsEmail, fixtures.AuthCredsPassword)

		s.Require().ErrorIs(err, saveErr)
		s.Nil(tokens)
	})
}
