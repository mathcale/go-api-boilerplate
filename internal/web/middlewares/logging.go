package middlewares

import (
	"net/http"
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type loggingMiddleware struct {
	logger logger.Logger
}

func NewLoggingMiddleware(logger logger.Logger) MiddlewareHandler {
	return &loggingMiddleware{
		logger: logger,
	}
}

func (m *loggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			m.logger.Info("Incoming request", map[string]interface{}{
				"method":     r.Method,
				"url":        r.URL.RequestURI(),
				"user_agent": r.UserAgent(),
				"elapsed_ms": time.Since(start).Milliseconds(),
			})
		}()

		next.ServeHTTP(w, r)
	})
}
