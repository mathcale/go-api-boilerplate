package middlewares

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/clientip"
)

type (
	RateLimitMiddleware interface {
		Handler(next http.Handler) http.Handler
	}

	KeyLimiter interface {
		Allow(key string) bool
	}

	rateLimitMiddleware struct {
		response handlers.Response
		limiter  KeyLimiter
		resolver *clientip.Resolver
	}
)

func NewRateLimitMiddleware(
	limiter KeyLimiter,
	resolver *clientip.Resolver,
	response handlers.Response,
) RateLimitMiddleware {
	return &rateLimitMiddleware{
		response: response,
		limiter:  limiter,
		resolver: resolver,
	}
}

func (m *rateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.limiter.Allow(m.resolver.Resolve(r)) {
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
