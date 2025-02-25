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
	handlers       []handler
	pingHandler    handlers.Ping
	counterHandler handlers.Counter
}

func NewRouter(
	pingHandler handlers.Ping,
	counterHandler handlers.Counter,
) Router {
	return &router{
		pingHandler:    pingHandler,
		counterHandler: counterHandler,
	}
}

func (r *router) Handlers() []handler {
	r.setHealthRoutes()
	r.setExampleRoutes()
	// your routes here!

	return r.handlers
}

func (r *router) setHealthRoutes() {
	r.handlers = append(r.handlers, []handler{
		r.newHandler("/ping", http.MethodGet, r.pingHandler.Ping),
	}...)
}

func (r *router) setExampleRoutes() {
	r.handlers = append(r.handlers, []handler{
		r.newHandler("/v1/counter", http.MethodGet, r.counterHandler.Count),
	}...)
}

func (r *router) newHandler(
	path, method string,
	handlerFunc http.HandlerFunc,
) handler {
	return handler{
		path:        path,
		method:      method,
		handlerFunc: handlerFunc,
	}
}
