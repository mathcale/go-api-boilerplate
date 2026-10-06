package system_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	"github.com/mathcale/go-api-boilerplate/internal/modules/system"
)

func TestModule_Routes(t *testing.T) {
	mod := system.New(&core.Container{})

	assert.Equal(t, "system", mod.Name())

	routes := mod.Routes()
	require.Len(t, routes, 2)

	assert.Equal(t, http.MethodGet, routes[0].Method)
	assert.Equal(t, "/ping", routes[0].Path)
	assert.Equal(t, module.Public, routes[0].Protected)
	assert.Equal(t, module.RateLimited, routes[0].RateLimited)
	assert.Empty(t, routes[0].RequiredRoles)
	assert.NotNil(t, routes[0].Handler)

	assert.Equal(t, http.MethodGet, routes[1].Method)
	assert.Equal(t, "/v1/admin/ping", routes[1].Path)
	assert.Equal(t, module.Protected, routes[1].Protected)
	assert.Equal(t, module.NotThrottled, routes[1].RateLimited)
	assert.Equal(t, []string{"admin"}, routes[1].RequiredRoles)
	assert.NotNil(t, routes[1].Handler)
}
