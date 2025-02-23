package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/xid"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type loggingMiddleware struct {
	logger logger.Logger
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingMiddleware(logger logger.Logger) MiddlewareHandler {
	return &loggingMiddleware{
		logger: logger,
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
		correlationID := xid.New().String()

		ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
		r = r.WithContext(ctx)
		m.logger.SetGlobalValue("correlation_id", correlationID)

		w.Header().Add("X-Correlation-ID", correlationID)

		lrw := newLoggingResponseWriter(w)

		defer func() {
			panicVal := recover()
			if panicVal != nil {
				lrw.statusCode = http.StatusInternalServerError
				panic(panicVal)
			}

			m.logger.Info("Incoming request", map[string]interface{}{
				"method":      r.Method,
				"url":         r.URL.RequestURI(),
				"status_code": lrw.statusCode,
				"user_agent":  r.UserAgent(),
				"elapsed_ms":  time.Since(start),
			})
		}()

		next.ServeHTTP(lrw, r)
	})
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
