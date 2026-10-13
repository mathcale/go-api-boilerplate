package jwt_test

import (
	"testing"

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

	s.Run("should reject a malformed or tampered token string", func() {
		_, err := s.auth.VerifyAccessToken("not-a-real-token")

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

	s.Run("should reject an access token when verified as a refresh token", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
		s.Require().NoError(err)

		_, err = s.auth.VerifyRefreshToken(*issued.Token)

		s.Require().Error(err)
	})
}
