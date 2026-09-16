package middlewares

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

type (
	RateLimitMiddleware interface {
		Handler(next http.Handler) http.Handler
	}

	rateLimitMiddleware struct {
		response handlers.Response
		limit    rate.Limit
		burst    int
		mu       sync.Mutex
		limiters map[string]*rate.Limiter
	}
)

func NewRateLimitMiddleware(
	requestsPerMinute, burst int,
	response handlers.Response,
) RateLimitMiddleware {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}

	if burst <= 0 {
		burst = requestsPerMinute
	}

	return &rateLimitMiddleware{
		response: response,
		limit:    rate.Limit(float64(requestsPerMinute) / 60.0),
		burst:    burst,
		limiters: make(map[string]*rate.Limiter),
	}
}

func (m *rateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.limiterFor(m.clientIP(r)).Allow() {
			code := apperror.BE0005_RATE_LIMITED

			m.response.RespondWithError(w, apperror.New(
				nil, "rate limit exceeded", apperror.TooManyRequests,
				apperror.MiddlewareOrigin, "rate_limit", &code, nil,
			), map[string]string{"Retry-After": "60"})

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *rateLimitMiddleware) limiterFor(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, ok := m.limiters[ip]
	if !ok {
		limiter = rate.NewLimiter(m.limit, m.burst)
		m.limiters[ip] = limiter
	}

	return limiter
}

func (m *rateLimitMiddleware) clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}
