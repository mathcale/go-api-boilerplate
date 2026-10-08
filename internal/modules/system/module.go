package system

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	"github.com/mathcale/go-api-boilerplate/internal/modules/system/handlers"
)

type Module struct {
	ping handlers.PingHandler
}

func New(c *core.Container) *Module {
	return &Module{
		ping: handlers.NewPingHandler(c.Response),
	}
}

func (m *Module) Name() string {
	return "system"
}

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
			// FIXME: Demo of role-based access control where only users whose JWT carries the "admin" role may reach this route
			Method:        http.MethodGet,
			Path:          "/v1/admin/ping",
			Protected:     module.Protected,
			RateLimited:   module.NotThrottled,
			RequiredRoles: []string{"admin"},
			Handler:       m.ping.Handle,
		},
	}
}
