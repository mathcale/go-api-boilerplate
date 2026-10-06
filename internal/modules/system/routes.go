package system

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/core/module"
)

func (m *Module) Routes() []module.Route {
	return []module.Route{
		{
			Method:      http.MethodGet,
			Path:        "/ping",
			Protected:   module.Public,
			RateLimited: module.RateLimited,
			Handler:     m.ping.Handle,
		},
		{
			// FIXME: Demo of role-based access control where only users whose JWT
			// carries the "admin" role may reach this route
			Method:        http.MethodGet,
			Path:          "/v1/admin/ping",
			Protected:     module.Protected,
			RateLimited:   module.NotThrottled,
			RequiredRoles: []string{"admin"},
			Handler:       m.ping.Handle,
		},
	}
}
