package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/infra/web/handlers"
	"go-api-boilerplate/internal/infra/web/handlers/dto"
	"go-api-boilerplate/internal/infra/web/webcontext"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/tests/mocks"
)

type RequireRoleMiddlewareTestSuite struct {
	suite.Suite
}

func TestRequireRoleMiddleware(t *testing.T) {
	suite.Run(t, new(RequireRoleMiddlewareTestSuite))
}

func (s *RequireRoleMiddlewareTestSuite) TestHandler() {
	s.Run("should allow the request when the context has a matching role", func() {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		mw := s.newRoleMiddleware("admin")

		rec := httptest.NewRecorder()

		mw.Handler(next).ServeHTTP(rec, s.requestWithRoles([]string{"user", "admin"}))

		s.True(called)
		s.Equal(http.StatusOK, rec.Code)
	})

	s.Run("should forbid the request when the context is missing the required role", func() {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		mw := s.newRoleMiddleware("admin")

		rec := httptest.NewRecorder()

		mw.Handler(next).ServeHTTP(rec, s.requestWithRoles([]string{"user"}))

		s.False(called)
		s.assertForbidden(rec)
	})

	s.Run("should forbid the request when there are no roles in the context", func() {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.Fail("next handler should not be called")
		})

		mw := s.newRoleMiddleware("admin")

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil).
			WithContext(context.Background())
		rec := httptest.NewRecorder()

		mw.Handler(next).ServeHTTP(rec, req)

		s.assertForbidden(rec)
	})
}

func (s *RequireRoleMiddlewareTestSuite) newRoleMiddleware(
	accepted ...string,
) RequireRoleMiddleware {
	response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)

	return NewRequireRoleMiddleware(response, accepted...)
}

func (s *RequireRoleMiddlewareTestSuite) requestWithRoles(roles []string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	ctx := webcontext.WithRoles(req.Context(), roles)

	return req.WithContext(ctx)
}

func (s *RequireRoleMiddlewareTestSuite) assertForbidden(rec *httptest.ResponseRecorder) {
	s.Equal(http.StatusForbidden, rec.Code)

	var out dto.ErrorOutput
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))
	s.Require().NotNil(out.Code)
	s.Equal(apperror.BE0004_FORBIDDEN, *out.Code)
}
