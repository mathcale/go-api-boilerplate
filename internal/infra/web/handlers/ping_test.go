package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type PingHandlerTestSuite struct {
	suite.Suite
}

func TestPingHandler(t *testing.T) {
	suite.Run(t, new(PingHandlerTestSuite))
}

func (s *PingHandlerTestSuite) TestHandle() {
	s.Run("should return pong with a 200 status code", func() {
		response := handlers.NewResponse(mocks.NoopLogger{}, config.EnvironmentTest)
		h := handlers.NewPingHandler(response)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		h.Handle(rec, req)

		s.Equal(http.StatusOK, rec.Code)
		s.Equal("pong", rec.Body.String())
	})
}
