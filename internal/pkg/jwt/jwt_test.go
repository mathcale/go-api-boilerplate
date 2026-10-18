package jwt_test

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type JWTAuthTestSuite struct {
	suite.Suite
	auth jwt.JWTAuth
}

func (s *JWTAuthTestSuite) SetupTest() {
	s.auth = fixtures.DomainPkgJWTAuth()
}

func (s *JWTAuthTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestJWTAuth(t *testing.T) {
	suite.Run(t, new(JWTAuthTestSuite))
}

func (s *JWTAuthTestSuite) TestIssueAccessToken() {
	s.Run("should issue an access token with subject and issuer", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{
			UserID:      "user-123",
			ExtraClaims: jwt.ExtraClaims{Roles: []string{"admin", "user"}},
		})

		s.Require().NoError(err)
		s.Require().NotNil(issued.Token)
		s.Equal("user-123", issued.Subject)
		s.Equal(fixtures.DomainPkgIssuer, issued.Issuer)
	})

	s.Run("should issue an access token carrying its type and token version", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{
			UserID:       "user-123",
			TokenVersion: fixtures.DomainPkgTokenVersion,
		})

		s.Require().NoError(err)
		s.Equal(jwt.TokenTypeAccess, issued.Type)
		s.Equal(fixtures.DomainPkgTokenVersion, issued.TokenVersion)
	})

	s.Run("should issue an access token with no token ID by default", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})

		s.Require().NoError(err)
		s.Empty(issued.ID)
	})
}

func (s *JWTAuthTestSuite) TestIssueRefreshToken() {
	s.Run("should issue a refresh token carrying its token ID", func() {
		issued, err := s.auth.IssueRefreshToken(jwt.IssueTokenParams{
			UserID:  "user-123",
			TokenID: "some-uuid",
		})

		s.Require().NoError(err)
		s.Require().NotNil(issued.Token)
		s.Equal("some-uuid", issued.ID)
	})

	s.Run("should issue a refresh token carrying its type and token version", func() {
		issued, err := s.auth.IssueRefreshToken(jwt.IssueTokenParams{
			UserID:       "user-123",
			TokenVersion: fixtures.DomainPkgTokenVersion,
		})

		s.Require().NoError(err)
		s.Equal(jwt.TokenTypeRefresh, issued.Type)
		s.Equal(fixtures.DomainPkgTokenVersion, issued.TokenVersion)
	})
}

func (s *JWTAuthTestSuite) TestVerifyAccessToken() {
	s.Run("should verify an issued access token, preserving its claims", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{
			UserID:      "user-123",
			ExtraClaims: jwt.ExtraClaims{Roles: []string{"admin", "user"}},
		})
		s.Require().NoError(err)
		s.Require().NotNil(issued.Token)

		verified, err := s.auth.VerifyAccessToken(*issued.Token)

		s.Require().NoError(err)
		s.Equal("user-123", verified.Subject)
		s.Equal(fixtures.DomainPkgIssuer, verified.Issuer)
		s.Equal([]string{"admin", "user"}, verified.ExtraClaims.Roles)
		s.Empty(verified.ID)
	})

	s.Run("should round-trip the token type and version", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{
			UserID:       "user-123",
			TokenVersion: fixtures.DomainPkgTokenVersion,
		})
		s.Require().NoError(err)

		verified, err := s.auth.VerifyAccessToken(*issued.Token)

		s.Require().NoError(err)
		s.Equal(jwt.TokenTypeAccess, verified.Type)
		s.Equal(fixtures.DomainPkgTokenVersion, verified.TokenVersion)
		s.Equal([]string{fixtures.DomainPkgAudience}, verified.Audience)
	})

	s.Run("should accept a token whose aud claim is an array", func() {
		claims := fixtures.DomainPkgClaims()
		claims["aud"] = []string{fixtures.DomainPkgAudience, "other"}

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		verified, err := s.auth.VerifyAccessToken(tokenStr)

		s.Require().NoError(err)
		s.Equal([]string{fixtures.DomainPkgAudience, "other"}, verified.Audience)
	})

	s.Run("should reject a refresh token even when both secrets are identical", func() {
		shared := fixtures.DomainPkgSharedSecretJWTAuth()

		issued, err := shared.IssueRefreshToken(jwt.IssueTokenParams{UserID: "user-123"})
		s.Require().NoError(err)

		_, err = shared.VerifyAccessToken(*issued.Token)

		s.Require().Error(err)
	})

	s.Run("should reject a token without a type claim", func() {
		claims := fixtures.DomainPkgClaims()
		delete(claims, "typ")

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})

	s.Run("should reject a token with an unknown type claim", func() {
		claims := fixtures.DomainPkgClaims()
		claims["typ"] = "id"

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})

	s.Run("should reject a token without a subject instead of panicking", func() {
		claims := fixtures.DomainPkgClaims()
		delete(claims, "sub")

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		s.NotPanics(func() {
			_, err = s.auth.VerifyAccessToken(tokenStr)
		})
		s.Require().Error(err)
	})

	s.Run("should reject a token without an expiry instead of panicking", func() {
		claims := fixtures.DomainPkgClaims()
		delete(claims, "exp")

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		s.NotPanics(func() {
			_, err = s.auth.VerifyAccessToken(tokenStr)
		})
		s.Require().Error(err)
	})

	s.Run("should reject a token without a token version", func() {
		claims := fixtures.DomainPkgClaims()
		delete(claims, "tv")

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})

	s.Run("should reject a token with a non-numeric or negative token version", func() {
		for _, bad := range []any{"1", -1, 1.5} {
			claims := fixtures.DomainPkgClaims()
			claims["tv"] = bad

			tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
			s.Require().NoError(err)

			_, err = s.auth.VerifyAccessToken(tokenStr)

			s.Require().Error(err)
		}
	})

	s.Run("should reject a malformed or tampered token string", func() {
		_, err := s.auth.VerifyAccessToken("not-a-real-token")

		s.Require().Error(err)
	})

	s.Run("should reject a token issued by a different issuer", func() {
		claims := fixtures.DomainPkgClaims()
		claims["iss"] = "someone-else"

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().ErrorIs(err, jwtlib.ErrTokenInvalidIssuer)
	})

	s.Run("should reject a token issued for a different audience", func() {
		claims := fixtures.DomainPkgClaims()
		claims["aud"] = "other-clients"

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().ErrorIs(err, jwtlib.ErrTokenInvalidAudience)
	})

	s.Run("should reject a token with an invalid iat claim", func() {
		claims := fixtures.DomainPkgClaims()
		claims["iat"] = "not-a-timestamp"
		claims["exp"] = time.Now().Add(time.Hour).Unix()

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})

	s.Run("should reject a token with an invalid exp claim", func() {
		claims := fixtures.DomainPkgClaims()
		claims["exp"] = "not-a-timestamp"

		tokenStr, err := fixtures.DomainPkgSignAccessClaims(claims)
		s.Require().NoError(err)

		_, err = s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})
}

func (s *JWTAuthTestSuite) TestVerifyRefreshToken() {
	s.Run("should verify an issued refresh token, preserving its token ID", func() {
		issued, err := s.auth.IssueRefreshToken(jwt.IssueTokenParams{
			UserID:  "user-123",
			TokenID: "some-uuid",
		})
		s.Require().NoError(err)
		s.Require().NotNil(issued.Token)

		verified, err := s.auth.VerifyRefreshToken(*issued.Token)

		s.Require().NoError(err)
		s.Equal("some-uuid", verified.ID)
	})

	s.Run("should round-trip the token type and version", func() {
		issued, err := s.auth.IssueRefreshToken(jwt.IssueTokenParams{
			UserID:       "user-123",
			TokenVersion: fixtures.DomainPkgTokenVersion,
		})
		s.Require().NoError(err)

		verified, err := s.auth.VerifyRefreshToken(*issued.Token)

		s.Require().NoError(err)
		s.Equal(jwt.TokenTypeRefresh, verified.Type)
		s.Equal(fixtures.DomainPkgTokenVersion, verified.TokenVersion)
	})

	s.Run("should reject an access token even when both secrets are identical", func() {
		shared := fixtures.DomainPkgSharedSecretJWTAuth()

		issued, err := shared.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
		s.Require().NoError(err)

		_, err = shared.VerifyRefreshToken(*issued.Token)

		s.Require().Error(err)
	})

	s.Run("should reject an access token when verified as a refresh token", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
		s.Require().NoError(err)

		_, err = s.auth.VerifyRefreshToken(*issued.Token)

		s.Require().Error(err)
	})
}
