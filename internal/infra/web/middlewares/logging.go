package middlewares

import (
	"net/http"
	"time"

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

		lrw := newLoggingResponseWriter(w)

		defer func() {
			panicVal := recover()
			if panicVal != nil {
				lrw.statusCode = http.StatusInternalServerError
				panic(panicVal)
			}

			m.logger.Info("Finished request", map[string]interface{}{
				"method":      r.Method,
				"url":         r.URL.RequestURI(),
				"status_code": lrw.statusCode,
				"user_agent":  r.UserAgent(),
				"time":        time.Now().Format(time.RFC3339),
				"elapsed_ms":  time.Since(start),
			})
		}()

		m.logger.Info("Incoming request", map[string]interface{}{
			"method":      r.Method,
			"url":         r.URL.RequestURI(),
			"status_code": lrw.statusCode,
			"user_agent":  r.UserAgent(),
			"time":        time.Now().Format(time.RFC3339),
		})

		next.ServeHTTP(lrw, r)
	})
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
