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

func storedRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	return &user.RefreshToken{
		ID:        familyID,
		UserID:    userID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

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

func TestRefreshTokenUseCase(t *testing.T) {
	suite.Run(t, new(RefreshTokenUseCaseTestSuite))
}

func (s *RefreshTokenUseCaseTestSuite) TestSuccess() {
	s.Run("should issue a new token pair when the refresh token is valid", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)

		access := "new-access-token"
		refresh := "new-refresh-token"
		expiresAt := time.Now().Add(24 * time.Hour)

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(nil)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().NoError(err)
		s.Require().NotNil(tokens)
		s.Equal(access, tokens.AccessToken)
		s.Equal(refresh, tokens.RefreshToken)
		s.gw.AssertExpectations(s.T())
		s.j.AssertExpectations(s.T())
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestInvalidJWT() {
	s.Run("should return unauthorized when the refresh token JWT is invalid", func() {
		s.j.On("VerifyRefreshToken", "bad-token").Return(nil, errors.New("invalid"))

		tokens, err := s.uc.Execute(context.Background(), "bad-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.gw.AssertNotCalled(s.T(), "GetRefreshToken", mock.Anything, mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestUnknownTokenID() {
	s.Run("should return unauthorized when the token family is not found", func() {
		userID := uuid.New()
		familyID := uuid.New()

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(nil, nil)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestRevokedFamily() {
	s.Run("should return unauthorized when the token family was revoked", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		revokedAt := time.Now()
		stored.RevokedAt = &revokedAt

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.gw.AssertNotCalled(s.T(), "RevokeRefreshTokenFamily", mock.Anything, mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestReuseDetected() {
	s.Run("should revoke the token family and log a warning when reuse is detected", func() {
		log := new(mocks.Logger)
		uc := auth.NewRefreshTokenUseCase(log, s.j, s.gw)

		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		usedAt := time.Now()
		stored.UsedAt = &usedAt

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(nil)
		log.On("Warn", mock.Anything, mock.MatchedBy(func(tags map[string]interface{}) bool {
			return tags["business_code"] == apperror.BE1011_REFRESH_TOKEN_REUSE_DETECTED
		})).Return()

		tokens, err := uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		// The client-facing code stays generic: reuse detection must not be
		// distinguishable from any other unauthorized refresh via the API response.
		s.Equal(apperror.BE0003_UNAUTHORIZED, *appErr.BusinessCode())
		s.gw.AssertExpectations(s.T())
		log.AssertExpectations(s.T())
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestReuseDetected_RevokeFails() {
	s.Run("should log an error when revoking a reused token family fails", func() {
		log := new(mocks.Logger)
		uc := auth.NewRefreshTokenUseCase(log, s.j, s.gw)

		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		usedAt := time.Now()
		stored.UsedAt = &usedAt
		revokeErr := errors.New("db down")

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("RevokeRefreshTokenFamily", mock.Anything, familyID).Return(revokeErr)
		log.On("Error", mock.Anything, revokeErr, mock.Anything).Return()

		tokens, err := uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		log.AssertExpectations(s.T())
		log.AssertNotCalled(s.T(), "Warn", mock.Anything, mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestMissingJTI() {
	s.Run("should return unauthorized when the JWT is missing a token ID", func() {
		userID := uuid.New()

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: ""}, nil)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.gw.AssertNotCalled(s.T(), "GetRefreshToken", mock.Anything, mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestInvalidTokenID() {
	s.Run("should return unauthorized when the token ID is not a valid UUID", func() {
		userID := uuid.New()

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: "not-a-uuid"}, nil)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.UnauthorizedKind, appErr.Kind())
		s.gw.AssertNotCalled(s.T(), "GetRefreshToken", mock.Anything, mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestGetRefreshTokenGatewayError() {
	s.Run("should propagate the error when fetching the refresh token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		gwErr := errors.New("db down")

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(nil, gwErr)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().ErrorIs(err, gwErr)
		s.Nil(tokens)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestMarkRefreshTokenUsedError() {
	s.Run("should propagate the error when marking the refresh token as used fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		markErr := errors.New("db down")

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(markErr)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().ErrorIs(err, markErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueAccessToken", mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestIssueAccessTokenError() {
	s.Run("should propagate the error when issuing the access token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		issueErr := errors.New("signing failure")

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.j.AssertNotCalled(s.T(), "IssueRefreshToken", mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestIssueRefreshTokenError() {
	s.Run("should propagate the error when issuing the refresh token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		access := "new-access-token"
		issueErr := errors.New("signing failure")

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(nil, issueErr)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().ErrorIs(err, issueErr)
		s.Nil(tokens)
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestNewRefreshTokenMissingExpiry() {
	s.Run("should return a dependency error when the issued refresh token has no expiry", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		access := "new-access-token"
		refresh := "new-refresh-token"

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &refresh, ExpiresAt: nil}, nil)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().Error(err)
		s.Nil(tokens)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.DependencyKind, appErr.Kind())
		s.gw.AssertNotCalled(s.T(), "SaveRefreshToken", mock.Anything, mock.Anything)
	})
}

func (s *RefreshTokenUseCaseTestSuite) TestSaveRefreshTokenError() {
	s.Run("should propagate the error when saving the new refresh token fails", func() {
		userID := uuid.New()
		familyID := uuid.New()
		stored := storedRefreshToken(userID, familyID)
		access := "new-access-token"
		refresh := "new-refresh-token"
		expiresAt := time.Now().Add(24 * time.Hour)
		saveErr := errors.New("db down")

		s.j.On("VerifyRefreshToken", "old-refresh-token").
			Return(&jwt.Token{Subject: userID.String(), ID: familyID.String()}, nil)
		s.gw.On("GetRefreshToken", mock.Anything, familyID).Return(stored, nil)
		s.gw.On("MarkRefreshTokenUsed", mock.Anything, familyID).Return(nil)
		s.j.On("IssueAccessToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &access}, nil)
		s.j.On("IssueRefreshToken", mock.AnythingOfType("jwt.IssueTokenParams")).
			Return(&jwt.Token{Token: &refresh, ExpiresAt: &expiresAt}, nil)
		s.gw.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("user.RefreshToken")).
			Return(saveErr)

		tokens, err := s.uc.Execute(context.Background(), "old-refresh-token")

		s.Require().ErrorIs(err, saveErr)
		s.Nil(tokens)
	})
}
