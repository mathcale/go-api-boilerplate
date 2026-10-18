package web

import (
	"context"
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

	shutdownTimeout = 10 * time.Second
)

type Server interface {
	Start(ctx context.Context) error
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
	docsEnabled         bool
}

func NewServer(
	l logger.Logger,
	port int,
	docsEnabled bool,
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
		docsEnabled:         docsEnabled,
	}
}

func (s *server) Start(ctx context.Context) error {
	for _, h := range s.handlers {
		s.logger.Debug(
			"Registering route",
			"method", h.method,
			"path", h.path,
			"protected", h.protected,
			"rate_limited", h.rateLimited,
		)

		s.router.Handle(fmt.Sprintf("%s %s", h.method, h.path), s.wrap(h))
	}

	s.registerDocs()

	middlewareChain := alice.New()

	for _, mw := range s.middlewares {
		s.logger.Debug("Registering middleware", "name", mw.name)

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

	s.logger.Info("Starting http server", "port", s.port)

	errCh := make(chan error, 1)

	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("Shutting down http server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	}
}

func (s *server) registerDocs() {
	if !s.docsEnabled {
		return
	}

	s.router.Handle("GET /swagger/", httpSwagger.WrapHandler)
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
