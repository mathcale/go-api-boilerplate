package jwt_test

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
)

// accessSecret must match the secret used by newAuth() in jwt_test.go so these
// hand-crafted tokens are accepted by the signature check, letting us reach the
// claim-parsing branches inside toToken/parseTime that IssueAccessToken can
// never produce on its own (it always emits well-formed iat/exp claims).
const accessSecret = "access-secret"

func signWithClaims(t *testing.T, claims jwtlib.MapClaims) string {
	t.Helper()

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, claims)

	signed, err := token.SignedString([]byte(accessSecret))
	require.NoError(t, err)

	return signed
}

type JWTAuthClaimsTestSuite struct {
	suite.Suite
	auth jwt.JWTAuth
}

func (s *JWTAuthClaimsTestSuite) SetupTest() {
	s.auth = newAuth()
}

func TestJWTAuthClaims(t *testing.T) {
	suite.Run(t, new(JWTAuthClaimsTestSuite))
}

func (s *JWTAuthClaimsTestSuite) TestJWTAuth_VerifyRejectsInvalidIatClaim() {
	s.Run("should reject a token with an invalid iat claim", func() {
		tokenStr := signWithClaims(s.T(), jwtlib.MapClaims{
			"sub": "user-123",
			"iss": "boilerplate",
			"aud": "boilerplate-clients",
			"iat": "not-a-timestamp",
			"exp": time.Now().Add(time.Hour).Unix(),
		})

		_, err := s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})
}

func (s *JWTAuthClaimsTestSuite) TestJWTAuth_VerifyRejectsInvalidExpClaim() {
	s.Run("should reject a token with an invalid exp claim", func() {
		tokenStr := signWithClaims(s.T(), jwtlib.MapClaims{
			"sub": "user-123",
			"iss": "boilerplate",
			"aud": "boilerplate-clients",
			"iat": time.Now().Unix(),
			"exp": "not-a-timestamp",
		})

		_, err := s.auth.VerifyAccessToken(tokenStr)

		s.Require().Error(err)
	})
}
