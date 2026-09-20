package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

func newAuth() jwt.JWTAuth {
	return jwt.NewJWTAuth(
		mocks.NoopLogger{},
		[]byte("access-secret"),
		[]byte("refresh-secret"),
		15,
		60,
		"boilerplate",
		"boilerplate-clients",
	)
}

type JWTAuthTestSuite struct {
	suite.Suite
	auth jwt.JWTAuth
}

func (s *JWTAuthTestSuite) SetupTest() {
	s.auth = newAuth()
}

func TestJWTAuth(t *testing.T) {
	suite.Run(t, new(JWTAuthTestSuite))
}

func (s *JWTAuthTestSuite) TestJWTAuth_AccessTokenRoundTrip() {
	s.Run("should issue and verify an access token, preserving subject, issuer, and extra claims", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{
			UserID:      "user-123",
			ExtraClaims: jwt.ExtraClaims{Roles: []string{"admin", "user"}},
		})

		s.Require().NoError(err)
		s.Require().NotNil(issued.Token)

		verified, err := s.auth.VerifyAccessToken(*issued.Token)

		s.Require().NoError(err)
		s.Equal("user-123", verified.Subject)
		s.Equal("boilerplate", verified.Issuer)
		s.Equal([]string{"admin", "user"}, verified.ExtraClaims.Roles)
	})
}

func (s *JWTAuthTestSuite) TestJWTAuth_AccessTokenRejectedByRefreshVerifier() {
	s.Run("should reject an access token when verified as a refresh token", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
		s.Require().NoError(err)

		// A token signed with the access secret must not validate as a refresh token.
		_, err = s.auth.VerifyRefreshToken(*issued.Token)
		s.Require().Error(err)
	})
}

func (s *JWTAuthTestSuite) TestJWTAuth_RejectsTamperedToken() {
	s.Run("should reject a malformed or tampered token string", func() {
		_, err := s.auth.VerifyAccessToken("not-a-real-token")
		s.Require().Error(err)
	})
}

func (s *JWTAuthTestSuite) TestJWTAuth_RefreshTokenCarriesTokenID() {
	s.Run("should issue and verify a refresh token carrying its token ID", func() {
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
}

func (s *JWTAuthTestSuite) TestJWTAuth_AccessTokenHasNoTokenIDByDefault() {
	s.Run("should issue an access token with no token ID by default", func() {
		issued, err := s.auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
		s.Require().NoError(err)

		verified, err := s.auth.VerifyAccessToken(*issued.Token)

		s.Require().NoError(err)
		s.Empty(verified.ID)
	})
}
