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
	securityHeaders middlewares.MiddlewareHandler
	cors            middlewares.MiddlewareHandler
	bodyLimit       middlewares.MiddlewareHandler
	correlationID   middlewares.MiddlewareHandler
	logging         middlewares.MiddlewareHandler
}

func NewMiddlewaresResolver(
	securityHeaders middlewares.MiddlewareHandler,
	cors middlewares.MiddlewareHandler,
	bodyLimit middlewares.MiddlewareHandler,
	correlationID middlewares.MiddlewareHandler,
	logging middlewares.MiddlewareHandler,
) MiddlewaresResolver {
	return &middlewareResolver{
		securityHeaders: securityHeaders,
		cors:            cors,
		bodyLimit:       bodyLimit,
		correlationID:   correlationID,
		logging:         logging,
	}
}

func (mr *middlewareResolver) Resolve() []middlewareHandler {
	return []middlewareHandler{
		{
			name:        "security_headers",
			handlerFunc: mr.securityHeaders.Handler,
		},
		{
			name:        "cors",
			handlerFunc: mr.cors.Handler,
		},
		{
			name:        "body_limit",
			handlerFunc: mr.bodyLimit.Handler,
		},
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
