package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CORSMiddlewareTestSuite struct {
	suite.Suite
}

func TestCORSMiddleware(t *testing.T) {
	suite.Run(t, new(CORSMiddlewareTestSuite))
}

func (s *CORSMiddlewareTestSuite) serve(origins []string) *httptest.ResponseRecorder {
	s.T().Helper()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	NewCORSMiddleware(origins).Handler(next).ServeHTTP(rec, req)

	return rec
}

func (s *CORSMiddlewareTestSuite) TestHandler() {
	s.Run("should allow a configured origin", func() {
		rec := s.serve([]string{"https://example.com"})

		s.Equal(http.StatusOK, rec.Code)
		s.Equal("https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
		s.Equal("true", rec.Header().Get("Access-Control-Allow-Credentials"))
	})

	s.Run("should not allow an origin that is not configured", func() {
		rec := s.serve([]string{"https://other.com"})

		s.Empty(rec.Header().Get("Access-Control-Allow-Origin"))
	})

	s.Run("should emit no CORS headers when no origins are configured", func() {
		for _, origins := range [][]string{nil, {}} {
			rec := s.serve(origins)

			s.Equal(http.StatusOK, rec.Code)
			s.Empty(rec.Header().Get("Access-Control-Allow-Origin"))
			s.Empty(rec.Header().Get("Access-Control-Allow-Credentials"))
		}
	})
}
