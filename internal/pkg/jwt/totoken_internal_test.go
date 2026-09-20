package jwt

import (
	"testing"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

type ToTokenInternalTestSuite struct {
	suite.Suite
	auth *jwtAuth
}

func (s *ToTokenInternalTestSuite) SetupTest() {
	s.auth = &jwtAuth{}
}

func TestToTokenInternal(t *testing.T) {
	suite.Run(t, new(ToTokenInternalTestSuite))
}

// toToken's own exp/iat validation is normally unreachable through the public
// verify() path because jwtlib.Parse's default validator already rejects
// malformed "exp" claims before toToken ever runs. It is exercised directly
// here as a defensive-code safety net.
func (s *ToTokenInternalTestSuite) TestToToken_InvalidExpClaim() {
	s.Run("should return an error when the exp claim cannot be parsed", func() {
		claims := jwtlib.MapClaims{
			"sub": "user-123",
			"iss": "boilerplate",
			"aud": "boilerplate-clients",
			"iat": int64(1),
			"exp": "not-a-timestamp",
		}
		token := "irrelevant"

		_, err := s.auth.toToken(&token, claims)

		s.Require().Error(err)
	})
}
