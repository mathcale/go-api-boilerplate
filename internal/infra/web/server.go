package web

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type Server interface {
	Start() error
}

type server struct {
	logger         logger.Logger
	router         *http.ServeMux
	handlers       []handler
	middlewares    []middlewareHandler
	authMiddleware middlewares.MiddlewareHandler
	port           int
}

func NewServer(
	l logger.Logger,
	port int,
	handlers []handler,
	middlewares []middlewareHandler,
	authMiddleware middlewares.MiddlewareHandler,
) Server {
	return &server{
		logger:         l,
		router:         http.NewServeMux(),
		handlers:       handlers,
		middlewares:    middlewares,
		authMiddleware: authMiddleware,
		port:           port,
	}
}

func (s *server) Start() error {
	for _, h := range s.handlers {
		s.logger.Debug("Registering route", map[string]interface{}{
			"method":    h.method,
			"path":      h.path,
			"protected": h.protected,
		})

		var handler http.Handler = h.handlerFunc

		if h.protected {
			handler = s.authMiddleware.Handler(h.handlerFunc)
		}

		s.router.Handle(fmt.Sprintf("%s %s", h.method, h.path), handler)
	}

	middlewareChain := alice.New()

	for _, mw := range s.middlewares {
		s.logger.Debug("Registering middleware", map[string]interface{}{
			"name": mw.name,
		})

		middlewareChain = middlewareChain.Append(mw.handlerFunc)
	}

	s.logger.Info("Starting http server", map[string]interface{}{
		"port": s.port,
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), middlewareChain.Then(s.router))
}
