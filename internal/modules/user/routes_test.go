package user_test

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/config"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

func TestModule_Routes(t *testing.T) {
	sqlDB, dbMock, err := sqlmock.New()
	require.NoError(t, err)

	dbMock.ExpectClose()

	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	mod, err := user.New(&core.Container{
		Config: &config.Config{},
		Logger: mocks.NoopLogger{},
		DB:     sqlx.NewDb(sqlDB, "postgres"),
		JWT:    new(mocks.JWTAuth),
		Bcrypt: new(mocks.Password),
		Email:  new(mocks.EmailClient),
	})
	require.NoError(t, err)

	assert.Equal(t, "user", mod.Name())
	assert.NotNil(t, mod.Gateway())

	type expected struct {
		method    string
		path      string
		protected bool
	}

	want := []expected{
		{http.MethodPost, "/v1/auth/signup", module.Public},
		{http.MethodPost, "/v1/auth/signin", module.Public},
		{http.MethodPost, "/v1/auth/refresh-token", module.Public},
		{http.MethodPost, "/v1/auth/confirm-account", module.Public},
		{http.MethodPost, "/v1/auth/resend-confirmation-code", module.Public},
		{http.MethodPost, "/v1/auth/recovery-code", module.Public},
		{http.MethodPost, "/v1/auth/recovery-code/validate", module.Public},
		{http.MethodPut, "/v1/auth/password", module.Public},
		{http.MethodGet, "/v1/auth/me", module.Protected},
	}

	routes := mod.Routes()
	require.Len(t, routes, len(want))

	for i, w := range want {
		assert.Equal(t, w.method, routes[i].Method, w.path)
		assert.Equal(t, w.path, routes[i].Path)
		assert.Equal(t, w.protected, routes[i].Protected, w.path)
		assert.Equal(t, module.RateLimited, routes[i].RateLimited, w.path)
		assert.Empty(t, routes[i].RequiredRoles, w.path)
		assert.NotNil(t, routes[i].Handler, w.path)
	}
}
