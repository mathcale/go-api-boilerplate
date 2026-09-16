package middlewares

import (
	"net/http"
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type loggingMiddleware struct {
	logger logger.Logger
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingMiddleware(l logger.Logger) MiddlewareHandler {
	return &loggingMiddleware{
		logger: l,
	}
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		w,
		http.StatusOK,
	}
}

func (m *loggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if correlationID, ok := webcontext.CorrelationIDFromContext(r.Context()); ok {
			m.logger.SetGlobalValue("correlation_id", correlationID)
		}

		lrw := newLoggingResponseWriter(w)

		defer func() {
			if panicVal := recover(); panicVal != nil {
				lrw.statusCode = http.StatusInternalServerError

				m.logger.Error("Recovered from panic", nil, map[string]interface{}{
					"panic":  panicVal,
					"method": r.Method,
					"url":    r.URL.RequestURI(),
				})
			}

			m.logger.Info("Incoming request", map[string]interface{}{
				"method":      r.Method,
				"url":         r.URL.RequestURI(),
				"status_code": lrw.statusCode,
				"user_agent":  r.UserAgent(),
				"elapsed_ms":  time.Since(start).Milliseconds(),
			})
		}()

		next.ServeHTTP(lrw, r)
	})
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
