package middlewares

import "net/http"

const MaxRequestBodyBytes int64 = 1 << 20

type bodyLimitMiddleware struct {
	maxBytes int64
}

func NewBodyLimitMiddleware(maxBytes int64) MiddlewareHandler {
	return &bodyLimitMiddleware{maxBytes: maxBytes}
}

func (m *bodyLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, m.maxBytes)

		next.ServeHTTP(w, r)
	})
}
