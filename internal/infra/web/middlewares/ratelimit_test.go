package middlewares

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/ratelimit"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/clientip"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type RateLimitMiddlewareTestSuite struct {
	suite.Suite
}

func TestRateLimitMiddleware(t *testing.T) {
	suite.Run(t, new(RateLimitMiddlewareTestSuite))
}

func (s *RateLimitMiddlewareTestSuite) newHandler(trusted string, burst int) http.Handler {
	s.T().Helper()

	limiter := ratelimit.NewKeyedLimiter(ratelimit.Config{RequestsPerMinute: 60, Burst: burst})
	s.T().Cleanup(limiter.Close)

	prefixes, err := clientip.ParseTrustedProxies(trusted)
	s.Require().NoError(err)

	response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)

	return NewRateLimitMiddleware(limiter, clientip.NewResolver(prefixes), response).
		Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
}

func (s *RateLimitMiddlewareTestSuite) serve(
	h http.Handler,
	remote, xff string,
) *httptest.ResponseRecorder {
	s.T().Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remote

	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func (s *RateLimitMiddlewareTestSuite) TestHandler() {
	s.Run("should let requests through within the burst", func() {
		h := s.newHandler("", 2)

		s.Equal(http.StatusOK, s.serve(h, "203.0.113.1:1000", "").Code)
		s.Equal(http.StatusOK, s.serve(h, "203.0.113.1:1001", "").Code)
	})

	s.Run("should keep separate buckets per client address", func() {
		h := s.newHandler("", 1)

		s.Equal(http.StatusOK, s.serve(h, "203.0.113.1:1000", "").Code)
		s.Equal(http.StatusOK, s.serve(h, "203.0.113.2:1000", "").Code)
	})

	s.Run("should separate clients behind a trusted proxy by forwarded address", func() {
		h := s.newHandler("10.0.0.1", 1)

		s.Equal(http.StatusOK, s.serve(h, "10.0.0.1:80", "198.51.100.1").Code)
		s.Equal(http.StatusOK, s.serve(h, "10.0.0.1:80", "198.51.100.2").Code)
		s.Equal(http.StatusTooManyRequests, s.serve(h, "10.0.0.1:80", "198.51.100.1").Code)
	})

	s.Run("should not let an untrusted client rotate a spoofed forwarded header", func() {
		h := s.newHandler("10.0.0.1", 1)

		s.Equal(http.StatusOK, s.serve(h, "203.0.113.1:1000", "198.51.100.1").Code)
		s.Equal(http.StatusTooManyRequests, s.serve(h, "203.0.113.1:1000", "198.51.100.2").Code)
	})

	s.Run("should reject requests over the limit with a retry hint", func() {
		h := s.newHandler("", 1)
		s.serve(h, "203.0.113.1:1000", "")

		rec := s.serve(h, "203.0.113.1:1000", "")

		s.Equal(http.StatusTooManyRequests, rec.Code)
		s.Equal("60", rec.Header().Get("Retry-After"))
	})

	s.Run("should be safe under concurrent requests", func() {
		h := s.newHandler("", 5)

		var wg sync.WaitGroup

		for i := 0; i < 20; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				s.serve(h, "203.0.113.1:1000", "")
			}()
		}

		wg.Wait()
	})
}
