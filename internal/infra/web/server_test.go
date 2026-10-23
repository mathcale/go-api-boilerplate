package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
}

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) serveSwagger(docsEnabled bool) int {
	srv := &server{router: http.NewServeMux(), docsEnabled: docsEnabled}
	srv.registerDocs()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	srv.router.ServeHTTP(rec, req)

	return rec.Code
}

func (s *ServerTestSuite) TestRegisterDocs() {
	s.Run("should serve swagger when docs are enabled", func() {
		s.Equal(http.StatusOK, s.serveSwagger(true))
	})

	s.Run("should return not found when docs are disabled", func() {
		s.Equal(http.StatusNotFound, s.serveSwagger(false))
	})
}
