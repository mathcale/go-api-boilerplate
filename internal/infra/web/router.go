package web

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
)

type Router interface {
	Handlers() []handler
}

type handler struct {
	path        string
	method      string
	handlerFunc http.HandlerFunc
}

type router struct {
	pingHandler    handlers.PingHandler
	counterHandler handlers.CounterHandler
}

func NewRouter(
	pingHandler handlers.PingHandler,
	counterHandler handlers.CounterHandler,
) Router {
	return &router{
		pingHandler:    pingHandler,
		counterHandler: counterHandler,
	}
}

func (r *router) Handlers() []handler {
	return []handler{
		{
			path:        "/ping",
			method:      http.MethodGet,
			handlerFunc: r.pingHandler.Handle,
		},
		{
			path:        "/v1/counter",
			method:      http.MethodGet,
			handlerFunc: r.counterHandler.Handle,
		},
	}
}
