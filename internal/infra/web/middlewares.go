package web

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
)

type MiddlewaresResolver interface {
	Resolve() []middlewareHandler
}

type middlewareHandler struct {
	name        string
	handlerFunc func(http.Handler) http.Handler
}

type middlewareResolver struct {
	correlationID middlewares.MiddlewareHandler
	logging       middlewares.MiddlewareHandler
}

func NewMiddlewaresResolver(
	correlationID middlewares.MiddlewareHandler,
	logging middlewares.MiddlewareHandler,
) MiddlewaresResolver {
	return &middlewareResolver{
		correlationID: correlationID,
		logging:       logging,
	}
}

func (mr *middlewareResolver) Resolve() []middlewareHandler {
	return []middlewareHandler{
		{
			name:        "correlation_id",
			handlerFunc: mr.correlationID.Handler,
		},
		{
			name:        "logging",
			handlerFunc: mr.logging.Handler,
		},
	}
}
