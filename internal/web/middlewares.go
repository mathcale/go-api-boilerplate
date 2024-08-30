package web

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/web/middlewares"
)

type MiddlewaresResolver interface {
	Resolve() []middlewareHandler
}

type middlewareHandler struct {
	name        string
	handlerFunc func(http.Handler) http.Handler
}

type middlewareResolver struct {
	logging middlewares.MiddlewareHandler
}

func NewMiddlewaresResolver(
	logging middlewares.MiddlewareHandler,
) MiddlewaresResolver {
	return &middlewareResolver{
		logging: logging,
	}
}

func (mr *middlewareResolver) Resolve() []middlewareHandler {
	return []middlewareHandler{
		{
			name:        "logging",
			handlerFunc: mr.logging.Handler,
		},
	}
}
