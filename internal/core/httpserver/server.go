package httpserver

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinas/alice"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/core/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	"github.com/mathcale/go-api-boilerplate/internal/core/response"
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
	response            response.Response
	routes              []module.Route
	middlewares         []middlewareHandler
	authMiddleware      middlewares.AuthMiddleware
	rateLimitMiddleware middlewares.RateLimitMiddleware
	port                int
}

func New(c *core.Container, routes []module.Route) Server {
	globalMiddlewares := NewMiddlewaresResolver(
		middlewares.NewSecurityHeadersMiddleware(),
		middlewares.NewCORSMiddleware(corsAllowedOrigins(c.Config.CORSAllowedOrigins)),
		middlewares.NewCorrelationIDMiddleware(c.Logger),
		middlewares.NewLoggingMiddleware(c.Logger),
	).Resolve()

	return &server{
		logger:         c.Logger,
		router:         http.NewServeMux(),
		response:       c.Response,
		routes:         routes,
		middlewares:    globalMiddlewares,
		authMiddleware: middlewares.NewAuthMiddleware(c.JWT, c.Response),
		rateLimitMiddleware: middlewares.NewRateLimitMiddleware(
			c.Config.RateLimitRequestsPerMinute,
			c.Config.RateLimitBurst,
			c.Response,
		),
		port: c.Config.WebServerPort,
	}
}

func corsAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	origins := strings.Split(raw, ",")

	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return origins
}

func (s *server) Start() error {
	for _, r := range s.routes {
		s.logger.Debug("Registering route", map[string]interface{}{
			"method":       r.Method,
			"path":         r.Path,
			"protected":    r.Protected,
			"rate_limited": r.RateLimited,
		})

		s.router.Handle(fmt.Sprintf("%s %s", r.Method, r.Path), s.wrap(r))
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

func (s *server) wrap(r module.Route) http.Handler {
	var wrapped http.Handler = r.Handler

	if len(r.RequiredRoles) > 0 {
		wrapped = middlewares.
			NewRequireRoleMiddleware(s.response, r.RequiredRoles...).
			Handler(wrapped)
	}

	if r.Protected {
		wrapped = s.authMiddleware.Handler(wrapped)
	}

	if r.RateLimited {
		wrapped = s.rateLimitMiddleware.Handler(wrapped)
	}

	return wrapped
}
