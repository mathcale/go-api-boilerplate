package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

func requestWithRoles(roles []string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	ctx := webcontext.WithRoles(req.Context(), roles)
	return req.WithContext(ctx)
}

func newRoleMiddleware(accepted ...string) middlewares.RequireRoleMiddleware {
	response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)
	return middlewares.NewRequireRoleMiddleware(response, accepted...)
}

func TestRequireRoleMiddleware_AllowsMatchingRole(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	mw := newRoleMiddleware("admin")

	rec := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, requestWithRoles([]string{"user", "admin"}))

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRoleMiddleware_ForbidsMissingRole(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	mw := newRoleMiddleware("admin")

	rec := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, requestWithRoles([]string{"user"}))

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRoleMiddleware_ForbidsWhenNoRolesInContext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	mw := newRoleMiddleware("admin")

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil).WithContext(context.Background())
	rec := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}
