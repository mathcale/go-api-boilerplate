package core_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
)

type stubModule struct {
	name   string
	routes []module.Route
}

func (s stubModule) Name() string           { return s.name }
func (s stubModule) Routes() []module.Route { return s.routes }

func TestCollectRoutes_ConcatenatesInOrder(t *testing.T) {
	first := stubModule{
		name: "first",
		routes: []module.Route{
			{Method: http.MethodGet, Path: "/a"},
			{Method: http.MethodPost, Path: "/b"},
		},
	}
	second := stubModule{
		name:   "second",
		routes: []module.Route{{Method: http.MethodPut, Path: "/c"}},
	}

	routes := core.CollectRoutes(first, second)

	paths := make([]string, 0, len(routes))
	for _, r := range routes {
		paths = append(paths, r.Path)
	}

	assert.Equal(t, []string{"/a", "/b", "/c"}, paths)
}

func TestCollectRoutes_NoModules(t *testing.T) {
	assert.Empty(t, core.CollectRoutes())
}
