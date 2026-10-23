package auth_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/pkg/jwt"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
	"go-api-boilerplate/internal/usecases/auth"
)

const oldRefreshToken = "old-refresh-token"

type RefreshTokenUseCaseTestSuite struct {
	suite.Suite
	gw *mocks.UserGateway
	j  *mocks.JWTAuth
	uc auth.RefreshTokenUseCase
}

func (s *RefreshTokenUseCaseTestSuite) SetupTest() {
	s.gw = new(mocks.UserGateway)
	s.j = new(mocks.JWTAuth)
	s.uc = auth.NewRefreshTokenUseCase(mocks.NoopLogger{}, s.j, s.gw)
}

func (s *RefreshTokenUseCaseTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestRefreshTokenUseCase(t *testing.T) {
	suite.Run(t, new(RefreshTokenUseCaseTestSuite))
}

func (s *RefreshTokenUseCaseTestSuite) expectVerified(userID uuid.UUID, tokenID string) {
	s.j.On("VerifyRefreshToken", oldRefreshToken).
		Return(fixtures.AuthSessionVerifiedJWT(userID, tokenID), nil)
}

func (s *RefreshTokenUseCaseTestSuite) expectRotationStart(
	userID, familyID uuid.UUID,
	stored *user.RefreshToken,
) {
	s.expectVerified(userID, familyID.String())
	s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
	s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(true, nil)
	s.gw.On("GetUserByID", mock.Anything, userID).Return(fixtures.AuthSessionUser(userID, true), nil)
}

func (s *RefreshTokenUseCaseTestSuite) assertUnauthorized(err error) {
	s.T().Helper()

	var appErr apperror.AppError
	s.Require().ErrorAs(err, &appErr)
	s.Equal(apperror.UnauthorizedKind, appErr.Kind())
	s.Require().NotNil(appErr.BusinessCode())
	s.Equal(apperror.BE0003_UNAUTHORIZED, *appErr.BusinessCode())
}

func (s *RefreshTokenUseCaseTestSuite) TestExecute() {
	s.Run("should issue a new token pair when the refresh token is valid", func() {
		userID := uuid.New()
		familyID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)

		s.expectRotationStart(userID, familyID, fixtures.AuthSessionStoredRefreshToken(userID, familyID))
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionAccessJWT("new-access-token"), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionRefreshJWT("new-refresh-token", &expiresAt), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().NoError(err)
		s.Require().NotNil(tokens)
		s.Equal("new-access-token", tokens.AccessToken)
		s.Equal("new-refresh-token", tokens.RefreshToken)
		s.gw.AssertExpectations(s.T())
		s.j.AssertExpectations(s.T())
	})

	s.Run("should issue tokens carrying the user's current token version", func() {
		userID := uuid.New()
		familyID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)
		hasVersion := mock.MatchedBy(func(p jwt.IssueTokenParams) bool {
			return p.TokenVersion == 3
		})

		s.j.On("VerifyRefreshToken", oldRefreshToken).
			Return(fixtures.AuthSessionVerifiedJWTWithVersion(userID, familyID.String(), 3), nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(true, nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUserWithVersion(userID, 3), nil)
		s.j.On("IssueAccessToken", hasVersion).
			Return(fixtures.AuthSessionAccessJWT("new-access-token"), nil)
		s.j.On("IssueRefreshToken", hasVersion).
			Return(fixtures.AuthSessionRefreshJWT("new-refresh-token", &expiresAt), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().NoError(err)
		s.Require().NotNil(tokens)
		s.j.AssertExpectations(s.T())
	})

	s.Run("should issue tokens with the roles currently stored for the user", func() {
		userID := uuid.New()
		familyID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)
		currentRoles := []string{"user"}
		hasCurrentRoles := mock.MatchedBy(func(p jwt.IssueTokenParams) bool {
			return p.UserID == userID.String() && slices.Equal(currentRoles, p.ExtraClaims.Roles)
		})

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(true, nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUserWithRoles(userID, currentRoles), nil)
		s.j.On("IssueAccessToken", hasCurrentRoles).
			Return(fixtures.AuthSessionAccessJWT("new-access-token"), nil)
		s.j.On("IssueRefreshToken", hasCurrentRoles).
			Return(fixtures.AuthSessionRefreshJWT("new-refresh-token", &expiresAt), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().NoError(err)
		s.Require().NotNil(tokens)
		s.j.AssertExpectations(s.T())
	})

	s.Run("should return unauthorized when the refresh token JWT is invalid", func() {
		s.j.On("VerifyRefreshToken", "bad-token").Return(nil, errors.New("invalid"))

		tokens, err := s.uc.Execute(context.Background(), "bad-token")

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertNotCalled(s.T(), "GetRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should return unauthorized when the JWT is missing a token ID", func() {
		s.expectVerified(uuid.New(), "")

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertNotCalled(s.T(), "GetRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should return unauthorized when the token ID is not a valid UUID", func() {
		s.expectVerified(uuid.New(), "not-a-uuid")

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertNotCalled(s.T(), "GetRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should return unauthorized when the token family is not found", func() {
		familyID := uuid.New()

		s.expectVerified(uuid.New(), familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(nil, nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should return unauthorized when the token family was revoked", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionRevokedRefreshToken(userID, familyID), nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertNotCalled(s.T(), "RevokeRefreshTokenFamily", mock.Anything, mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should revoke the token family and log a warning when reuse is detected", func() {
		log := new(mocks.Logger)
		uc := auth.NewRefreshTokenUseCase(log, s.j, s.gw)
		userID := uuid.New()
		familyID := uuid.New()

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionUsedRefreshToken(userID, familyID), nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(nil)
		log.On("Warn", mock.Anything, mock.MatchedBy(func(kv []any) bool {
			i := slices.Index(kv, any("business_code"))

			return i >= 0 && i+1 < len(kv) &&
				kv[i+1] == apperror.BE1011_REFRESH_TOKEN_REUSE_DETECTED
		})).Return()

		tokens, err := uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertExpectations(s.T())
		log.AssertExpectations(s.T())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})

	s.Run("should log an error when revoking a reused token family fails", func() {
		log := new(mocks.Logger)
		uc := auth.NewRefreshTokenUseCase(log, s.j, s.gw)
		userID := uuid.New()
		familyID := uuid.New()
		revokeErr := errors.New("db down")

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionUsedRefreshToken(userID, familyID), nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(revokeErr)
		log.On("Error", mock.Anything, revokeErr, mock.Anything).Return()

		tokens, err := uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		log.AssertExpectations(s.T())
		log.AssertNotCalled(s.T(), "Warn", mock.Anything, mock.Anything)
	})

	s.Run("should revoke the family and issue nothing when the rotation race is lost", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, true), nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(false, nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertExpectations(s.T())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when fetching the refresh token fails", func() {
		familyID := uuid.New()
		gwErr := errors.New("db down")

		s.expectVerified(uuid.New(), familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(nil, gwErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, gwErr)
		s.Nil(tokens)
	})

	s.Run("should propagate the error when marking the refresh token as used fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		markErr := errors.New("db down")

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, true), nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(false, markErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, markErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should return unauthorized when the token owner no longer exists", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should return unauthorized when the token owner is inactive", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUser(userID, false), nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should revoke the family and reject a refresh token with a stale version", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.j.On("VerifyRefreshToken", oldRefreshToken).
			Return(fixtures.AuthSessionVerifiedJWTWithVersion(userID, familyID.String(), 2), nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUserWithVersion(userID, 3), nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.gw.AssertExpectations(s.T())
		s.gw.AssertNotCalled(s.T(), "MarkRefreshTokenUsed", mock.Anything, mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})

	s.Run("should still reject a stale version when revoking the family fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		revokeErr := errors.New("db down")

		s.j.On("VerifyRefreshToken", oldRefreshToken).
			Return(fixtures.AuthSessionVerifiedJWTWithVersion(userID, familyID.String(), 2), nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).
			Return(fixtures.AuthSessionUserWithVersion(userID, 3), nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(revokeErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, revokeErr)
		s.Nil(tokens)
		s.assertUnauthorized(err)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should propagate the error when fetching the token owner fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		gwErr := errors.New("db down")

		s.expectVerified(userID, familyID.String())
		s.gw.On("GetRefreshToken", mock.Anything, familyID).
			Return(fixtures.AuthSessionStoredRefreshToken(userID, familyID), nil)
		s.gw.On("GetUserByID", mock.Anything, userID).Return(nil, gwErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, gwErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})

	s.Run("should propagate the error when issuing the access token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		issueErr := errors.New("signing failure")

		s.expectRotationStart(userID, familyID, fixtures.AuthSessionStoredRefreshToken(userID, familyID))
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})

	s.Run("should propagate the error when issuing the refresh token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		issueErr := errors.New("signing failure")

		s.expectRotationStart(userID, familyID, fixtures.AuthSessionStoredRefreshToken(userID, familyID))
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionAccessJWT("new-access-token"), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should return a dependency error when the new refresh token has no expiry", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.expectRotationStart(userID, familyID, fixtures.AuthSessionStoredRefreshToken(userID, familyID))
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionAccessJWT("new-access-token"), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionRefreshJWT("new-refresh-token", nil), nil)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.DependencyKind, appErr.Kind())
		s.Nil(appErr.BusinessCode())
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})

	s.Run("should propagate the error when saving the new refresh token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)
		saveErr := errors.New("db down")

		s.expectRotationStart(userID, familyID, fixtures.AuthSessionStoredRefreshToken(userID, familyID))
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionAccessJWT("new-access-token"), nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(fixtures.AuthSessionRefreshJWT("new-refresh-token", &expiresAt), nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(saveErr)

		tokens, err := s.uc.Execute(context.Background(), oldRefreshToken)

		s.Require().ErrorIs(err, saveErr)
		s.Nil(tokens)
	})
}
