package web

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/web/handlers/counter"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers/hello"
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
	helloHandler   hello.HelloHandler
	counterHandler counter.CounterHandler
}

func NewRouter(
	helloHandler hello.HelloHandler,
	counterHandler counter.CounterHandler,
) Router {
	return &router{
		helloHandler:   helloHandler,
		counterHandler: counterHandler,
	}
}

func (r *router) Handlers() []handler {
	return []handler{
		{
			path:        "/v1/hello",
			method:      http.MethodGet,
			handlerFunc: r.helloHandler.Handle,
		},
		{
			path:        "/v1/counter",
			method:      http.MethodGet,
			handlerFunc: r.counterHandler.Handle,
		},
	}
}
