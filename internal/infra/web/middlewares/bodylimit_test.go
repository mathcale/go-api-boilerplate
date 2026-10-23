package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BodyLimitMiddlewareTestSuite struct {
	suite.Suite
}

func TestBodyLimitMiddleware(t *testing.T) {
	suite.Run(t, new(BodyLimitMiddlewareTestSuite))
}

func (s *BodyLimitMiddlewareTestSuite) TestHandler() {
	var readErr error

	handler := NewBodyLimitMiddleware(16).Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, readErr = io.ReadAll(r.Body)
		}),
	)

	s.Run("should let a body within the limit be read in full", func() {
		readErr = nil
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small"))

		handler.ServeHTTP(httptest.NewRecorder(), req)

		s.NoError(readErr)
	})

	s.Run("should make reading a body over the limit fail", func() {
		readErr = nil
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 17)))

		handler.ServeHTTP(httptest.NewRecorder(), req)

		var maxErr *http.MaxBytesError
		s.ErrorAs(readErr, &maxErr)
	})
}
