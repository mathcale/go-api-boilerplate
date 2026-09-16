package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/justinas/alice"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second
)

type Server interface {
	Start() error
}

type server struct {
	logger              logger.Logger
	router              *http.ServeMux
	response            handlers.Response
	handlers            []handler
	middlewares         []middlewareHandler
	authMiddleware      middlewares.AuthMiddleware
	rateLimitMiddleware middlewares.RateLimitMiddleware
	port                int
}

func NewServer(
	l logger.Logger,
	port int,
	response handlers.Response,
	handlers []handler,
	globalMiddlewares []middlewareHandler,
	authMiddleware middlewares.AuthMiddleware,
	rateLimitMiddleware middlewares.RateLimitMiddleware,
) Server {
	return &server{
		logger:              l,
		router:              http.NewServeMux(),
		response:            response,
		handlers:            handlers,
		middlewares:         globalMiddlewares,
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		port:                port,
	}
}

func (s *server) Start() error {
	for _, h := range s.handlers {
		s.logger.Debug("Registering route", map[string]interface{}{
			"method":       h.method,
			"path":         h.path,
			"protected":    h.protected,
			"rate_limited": h.rateLimited,
		})

		s.router.Handle(fmt.Sprintf("%s %s", h.method, h.path), s.wrap(h))
	}

	s.router.Handle("GET /swagger/", httpSwagger.WrapHandler)

	middlewareChain := alice.New()

	for _, mw := range s.middlewares {
		s.logger.Debug("Registering middleware", map[string]interface{}{
			"name": mw.name,
		})

		middlewareChain = middlewareChain.Append(mw.handlerFunc)
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           middlewareChain.Then(s.router),
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	s.logger.Info("Starting http server", map[string]interface{}{
		"port": s.port,
	})

	return srv.ListenAndServe()
}

func (s *server) wrap(h handler) http.Handler {
	var wrapped http.Handler = h.handlerFunc

	if len(h.requiredRoles) > 0 {
		wrapped = middlewares.
			NewRequireRoleMiddleware(s.response, h.requiredRoles...).
			Handler(wrapped)
	}

	if h.protected {
		wrapped = s.authMiddleware.Handler(wrapped)
	}

	if h.rateLimited {
		wrapped = s.rateLimitMiddleware.Handler(wrapped)
	}

	return wrapped
}
