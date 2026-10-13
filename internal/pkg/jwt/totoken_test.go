package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type JWTAuthClaimsTestSuite struct {
	suite.Suite
	auth jwt.JWTAuth
}

func (s *JWTAuthClaimsTestSuite) SetupTest() {
	s.auth = fixtures.DomainPkgJWTAuth()
}

func (s *JWTAuthClaimsTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestJWTAuthClaims(t *testing.T) {
	suite.Run(t, new(JWTAuthClaimsTestSuite))
}

func (s *JWTAuthClaimsTestSuite) TestVerifyAccessToken() {
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
