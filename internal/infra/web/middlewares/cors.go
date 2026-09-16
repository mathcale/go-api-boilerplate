package middlewares

import (
	"net/http"

	"github.com/rs/cors"
)

type corsMiddleware struct {
	handler func(http.Handler) http.Handler
}

// NewCORSMiddleware configures CORS from the comma-separated allowed origins.
// An empty or "*" origin list allows any origin, which is convenient for local
// development but should be tightened in production.
func NewCORSMiddleware(allowedOrigins []string) MiddlewareHandler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
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
