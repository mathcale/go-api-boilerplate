package middlewares

import "net/http"

type securityHeadersMiddleware struct{}

// NewSecurityHeadersMiddleware sets a baseline of conservative security headers
// on every response. Projects can tighten the CSP as their front-end needs.
func NewSecurityHeadersMiddleware() MiddlewareHandler {
	return &securityHeadersMiddleware{}
}

func (m *securityHeadersMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		headers.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
