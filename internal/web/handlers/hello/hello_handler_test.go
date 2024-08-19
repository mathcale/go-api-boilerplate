package hello

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/web/handlers"
)

type HelloHandlerTestSuite struct {
	suite.Suite
	ResponseHandler handlers.ResponseHandler

	HelloHandler HelloHandler
}

func (s *HelloHandlerTestSuite) SetupTest() {
	s.ResponseHandler = handlers.NewResponseHandler()

	s.HelloHandler = NewHelloHandler(s.ResponseHandler)
}

func TestHelloHandler(t *testing.T) {
	suite.Run(t, new(HelloHandlerTestSuite))
}

func (s *HelloHandlerTestSuite) TestHandle() {
	s.Run("should return hello message", func() {
		r := httptest.NewRequest(http.MethodGet, "/hello", nil)
		w := httptest.NewRecorder()

		s.HelloHandler.Handle(w, r)

		res := w.Result()
		defer res.Body.Close()

		s.Equal(http.StatusOK, res.StatusCode)
		s.Contains(w.Body.String(), "Hello, World!")
	})
}
