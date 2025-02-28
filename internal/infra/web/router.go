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
	protected   bool
	handlerFunc http.HandlerFunc
}

type router struct {
	production     bool
	handlers       []handler
	pingHandler    handlers.Ping
	authHandler    handlers.Auth
	counterHandler handlers.Counter
}

const (
	PROTECTED_ROUTE bool = true
	PUBLIC_ROUTE    bool = false
)

func NewRouter(
	prod bool,
	pingHandler handlers.Ping,
	authHandler handlers.Auth,
	counterHandler handlers.Counter,
) Router {
	return &router{
		production:     prod,
		pingHandler:    pingHandler,
		authHandler:    authHandler,
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
	r.setAuthRoutes()
	r.setExampleRoutes()
	// your routes here!

	return r.handlers
}

func (r *router) setHealthRoutes() {
	r.handlers = append(r.handlers, []handler{
		r.newHandler("/ping", http.MethodGet, PUBLIC_ROUTE, r.pingHandler.Ping),
	}...)
}

func (r *router) setAuthRoutes() {
	r.handlers = append(r.handlers, []handler{
		r.newHandler("/v1/auth/signin", http.MethodPost, PUBLIC_ROUTE, r.authHandler.SignIn),
		r.newHandler("/v1/auth/signup", http.MethodPost, PUBLIC_ROUTE, r.authHandler.SignUp),
	}...)
}

func (r *router) setExampleRoutes() {
	r.handlers = append(r.handlers, []handler{
		r.newHandler("/v1/counter", http.MethodGet, PROTECTED_ROUTE, r.counterHandler.Count),
	}...)
}

func (r *router) newHandler(
	path, method string,
	protected bool,
	handlerFunc http.HandlerFunc,
) handler {
	return handler{
		path:        path,
		method:      method,
		protected:   protected,
		handlerFunc: handlerFunc,
	}
}
