package web

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

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
	production     bool
	handlers       []handler
	pingHandler    handlers.Ping
	counterHandler handlers.Counter
}

func NewRouter(
	prod bool,
	pingHandler handlers.Ping,
	counterHandler handlers.Counter,
) Router {
	return &router{
		production:     prod,
		pingHandler:    pingHandler,
		counterHandler: counterHandler,
	}
}

func (r *router) Handlers() []handler {
	if !r.production {
		r.handlers = append(
			r.handlers,
			r.newHandler("/swagger/", http.MethodGet, httpSwagger.Handler(
				httpSwagger.URL("/swagger/doc.json"),
				httpSwagger.DefaultModelsExpandDepth(httpSwagger.HideModel),
			)),
		)
	}

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
