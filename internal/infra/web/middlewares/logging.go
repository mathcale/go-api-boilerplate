package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"go-api-boilerplate/internal/infra/web/handlers/dto"
	"go-api-boilerplate/internal/pkg/logger"
)

type loggingMiddleware struct {
	logger logger.Logger
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func NewLoggingMiddleware(l logger.Logger) MiddlewareHandler {
	return &loggingMiddleware{
		logger: l,
	}
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (m *loggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := logger.FromCtxOr(r.Context(), m.logger)
		lrw := newLoggingResponseWriter(w)

		defer func() {
			if panicVal := recover(); panicVal != nil {
				log.Error(
					"Recovered from panic",
					fmt.Errorf("panic: %v", panicVal),
					"method", r.Method,
					"url", r.URL.RequestURI(),
					"stack", string(debug.Stack()),
				)

				if !lrw.headerWritten {
					lrw.respondInternalError()
				}
			}

			log.Info(
				"Incoming request",
				"method", r.Method,
				"url", r.URL.RequestURI(),
				"status_code", lrw.statusCode,
				"user_agent", r.UserAgent(),
				"elapsed_ms", time.Since(start).Milliseconds(),
			)
		}()

		next.ServeHTTP(lrw, r)
	})
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	if lrw.headerWritten {
		return
	}

	lrw.headerWritten = true
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if !lrw.headerWritten {
		lrw.WriteHeader(http.StatusOK)
	}

	return lrw.ResponseWriter.Write(b)
}

func (lrw *loggingResponseWriter) respondInternalError() {
	lrw.Header().Set("Content-Type", "application/json")
	lrw.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(lrw.ResponseWriter).Encode(
		dto.ErrorOutput{Message: "internal server error"},
	)
}
