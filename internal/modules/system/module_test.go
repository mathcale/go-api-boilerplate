package system_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	"github.com/mathcale/go-api-boilerplate/internal/modules/system"
)

type SystemModuleTestSuite struct {
	suite.Suite
}

func TestSystemModule(t *testing.T) {
	suite.Run(t, new(SystemModuleTestSuite))
}

func (s *SystemModuleTestSuite) TestName() {
	s.Run("should return module name", func() {
		mod := system.New(&core.Container{})

		s.Equal("system", mod.Name())
	})
}

func (s *SystemModuleTestSuite) TestRoutes() {
	s.Run("should register routes with middlewares and handler func", func() {
		mod := system.New(&core.Container{})
		routes := mod.Routes()

		s.Len(routes, 2)

		s.Equal(http.MethodGet, routes[0].Method)
		s.Equal("/ping", routes[0].Path)
		s.Equal(module.Public, routes[0].Protected)
		s.Equal(module.RateLimited, routes[0].RateLimited)
		s.Empty(routes[0].RequiredRoles)
		s.NotNil(routes[0].Handler)

		s.Equal(http.MethodGet, routes[1].Method)
		s.Equal("/v1/admin/ping", routes[1].Path)
		s.Equal(module.Protected, routes[1].Protected)
		s.Equal(module.NotThrottled, routes[1].RateLimited)
		s.Equal([]string{"admin"}, routes[1].RequiredRoles)
		s.NotNil(routes[1].Handler)
	})
}
