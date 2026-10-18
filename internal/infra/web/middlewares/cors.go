package middlewares

import (
	"net/http"

	"github.com/rs/cors"
)

type corsMiddleware struct {
	handler func(http.Handler) http.Handler
}

func NewCORSMiddleware(allowedOrigins []string) MiddlewareHandler {
	if len(allowedOrigins) == 0 {
		return &corsMiddleware{handler: func(next http.Handler) http.Handler { return next }}
	}

	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"Authorization", "Content-Type", correlationIDHeader},
		ExposedHeaders:   []string{correlationIDHeader},
		AllowCredentials: true,
	})

	return &corsMiddleware{handler: c.Handler}
}

func (m *corsMiddleware) Handler(next http.Handler) http.Handler {
	return m.handler(next)
}
