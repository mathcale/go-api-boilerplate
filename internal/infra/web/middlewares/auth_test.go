package middlewares

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/infra/web/handlers"
	"go-api-boilerplate/internal/infra/web/handlers/dto"
	"go-api-boilerplate/internal/infra/web/webcontext"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/tests/fixtures"
	"go-api-boilerplate/internal/tests/mocks"
)

const authTestBearer = "Bearer valid-token"

type AuthMiddlewareTestSuite struct {
	suite.Suite
	jwtAuth *mocks.JWTAuth
	users   *mocks.UserGateway
	mw      AuthMiddleware
}

func (s *AuthMiddlewareTestSuite) SetupTest() {
	s.jwtAuth = new(mocks.JWTAuth)
	s.users = new(mocks.UserGateway)

	response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)
	s.mw = NewAuthMiddleware(s.jwtAuth, s.users, response)
}

func (s *AuthMiddlewareTestSuite) SetupSubTest() {
	s.SetupTest()
}

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

func (s *AuthMiddlewareTestSuite) TestHandler() {
	s.Run("should pass the request with user and roles in context when the token is current", func() {
		userID := uuid.New()
		s.jwtAuth.On("VerifyAccessToken", "valid-token").
			Return(fixtures.WebVerifiedAccessToken(userID, 2), nil)
		s.users.On("GetUserAuthState", mock.Anything, userID).
			Return(fixtures.WebAuthState(2, true), nil)

		var gotID uuid.UUID
		var gotRoles []string
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			gotID, _ = webcontext.UserIDFromContext(r.Context())
			gotRoles, _ = webcontext.RolesFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		rec := s.serve(next)

		s.True(called)
		s.Equal(http.StatusOK, rec.Code)
		s.Equal(userID, gotID)
		s.Equal(fixtures.WebRoles(), gotRoles)
	})

	s.Run("should respond unauthorized when the token version is stale", func() {
		userID := uuid.New()
		s.jwtAuth.On("VerifyAccessToken", "valid-token").
			Return(fixtures.WebVerifiedAccessToken(userID, 1), nil)
		s.users.On("GetUserAuthState", mock.Anything, userID).
			Return(fixtures.WebAuthState(2, true), nil)

		s.assertRejected(s.serve(s.failingNext()))
	})

	s.Run("should respond unauthorized when the user is inactive", func() {
		userID := uuid.New()
		s.jwtAuth.On("VerifyAccessToken", "valid-token").
			Return(fixtures.WebVerifiedAccessToken(userID, 2), nil)
		s.users.On("GetUserAuthState", mock.Anything, userID).
			Return(fixtures.WebAuthState(2, false), nil)

		s.assertRejected(s.serve(s.failingNext()))
	})

	s.Run("should respond unauthorized when the user does not exist", func() {
		userID := uuid.New()
		s.jwtAuth.On("VerifyAccessToken", "valid-token").
			Return(fixtures.WebVerifiedAccessToken(userID, 2), nil)
		s.users.On("GetUserAuthState", mock.Anything, userID).Return(nil, nil)

		s.assertRejected(s.serve(s.failingNext()))
	})

	s.Run("should respond internal server error when the gateway fails", func() {
		userID := uuid.New()
		s.jwtAuth.On("VerifyAccessToken", "valid-token").
			Return(fixtures.WebVerifiedAccessToken(userID, 2), nil)
		s.users.On("GetUserAuthState", mock.Anything, userID).
			Return(nil, errors.New("db down"))

		rec := s.serve(s.failingNext())

		s.Equal(http.StatusInternalServerError, rec.Code)

		var out dto.ErrorOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
		s.Nil(out.Code)
	})

	s.Run("should respond unauthorized without a user lookup when the token is invalid", func() {
		s.jwtAuth.On("VerifyAccessToken", "valid-token").Return(nil, errors.New("bad token"))

		s.assertRejected(s.serve(s.failingNext()))
		s.users.AssertNotCalled(s.T(), "GetUserAuthState", mock.Anything, mock.Anything)
	})

	s.Run("should respond unauthorized when the authorization header is missing", func() {
		rec := httptest.NewRecorder()

		s.mw.Handler(s.failingNext()).ServeHTTP(
			rec, httptest.NewRequest(http.MethodGet, "/v1/me", nil),
		)

		s.assertRejected(rec)
		s.jwtAuth.AssertNotCalled(s.T(), "VerifyAccessToken", mock.Anything)
	})
}

func (s *AuthMiddlewareTestSuite) serve(next http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", authTestBearer)

	rec := httptest.NewRecorder()
	s.mw.Handler(next).ServeHTTP(rec, req)

	return rec
}

func (s *AuthMiddlewareTestSuite) failingNext() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Fail("next handler should not be called")
	})
}

func (s *AuthMiddlewareTestSuite) assertRejected(rec *httptest.ResponseRecorder) {
	s.Equal(http.StatusUnauthorized, rec.Code)

	var out dto.ErrorOutput
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
	s.Require().NotNil(out.Code)
	s.Equal(apperror.BE0003_UNAUTHORIZED, *out.Code)
}
