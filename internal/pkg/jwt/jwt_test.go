package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type JWTAuthTestSuite struct {
	suite.Suite
	loggerMock *mocks.Logger
	id         string
	a          jwt.JWTAuth
}

func (s *JWTAuthTestSuite) SetupTest() {
	s.loggerMock = new(mocks.Logger)
	s.id = "any-id"

	s.loggerMock.On("Debug", mock.Anything, mock.Anything)

	s.a = jwt.NewJWTAuth(
		s.loggerMock,
		[]byte("any-at-secret"),
		[]byte("any-rt-secret"),
		15,
		1440,
		"any-token-issuer",
		"any-token-audience",
	)
}

func (s *JWTAuthTestSuite) cleanMocks() {
	s.loggerMock.ExpectedCalls = nil
	s.loggerMock.Calls = nil
}

func TestJWTAuth(t *testing.T) {
	suite.Run(t, new(JWTAuthTestSuite))
}

func (s *JWTAuthTestSuite) TestIssueAccessToken() {
	s.Run("should issue token", func() {
		defer s.cleanMocks()

		t, err := s.a.IssueAccessToken(s.id)

		s.NoError(err)
		s.NotNil(t)
	})
}

func (s *JWTAuthTestSuite) TestVerifyAccessToken() {
	s.Run("should verify valid token", func() {
		defer s.cleanMocks()

		token, err := s.a.IssueAccessToken(s.id)
		if err != nil {
			s.FailNow(err.Error())
		}

		t, err := s.a.VerifyAccessToken(*token.Token)

		s.NoError(err)
		s.NotNil(t)
		s.Equal(s.id, t.Subject)
	})

	s.Run("should return error when checking blank token", func() {
		defer s.cleanMocks()

		t, err := s.a.VerifyAccessToken("")

		s.Error(err)
		s.ErrorContains(err, "token parsing failed")
		s.Nil(t)
	})
}

func (s *JWTAuthTestSuite) TestIssueRefreshToken() {
	s.Run("should issue token", func() {
		defer s.cleanMocks()

		t, err := s.a.IssueRefreshToken(s.id)

		s.NoError(err)
		s.NotNil(t)
	})
}

func (s *JWTAuthTestSuite) TestVerifyRefreshToken() {
	s.Run("should verify valid token", func() {
		defer s.cleanMocks()

		token, err := s.a.IssueRefreshToken(s.id)
		if err != nil {
			s.FailNow(err.Error())
		}

		t, err := s.a.VerifyRefreshToken(*token.Token)

		s.NoError(err)
		s.NotNil(t)
		s.Equal(s.id, t.Subject)
	})

	s.Run("should return error when checking blank token", func() {
		defer s.cleanMocks()

		t, err := s.a.VerifyRefreshToken("")

		s.Error(err)
		s.ErrorContains(err, "token parsing failed")
		s.Nil(t)
	})
}
