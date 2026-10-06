package system

import (
	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/modules/system/handlers"
)

type Module struct {
	ping handlers.PingHandler
}

func New(c *core.Container) *Module {
	return &Module{ping: handlers.NewPingHandler(c.Response)}
}

func (m *Module) Name() string {
	return "system"
}
