package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

type ResponseTestSuite struct {
	suite.Suite
}

func TestResponse(t *testing.T) {
	suite.Run(t, new(ResponseTestSuite))
}

func (s *ResponseTestSuite) TestRespondWithError() {
	cause := errors.New("db password leaked")

	respond := func(env string) dto.ErrorOutput {
		rec := httptest.NewRecorder()
		NewResponse(mocks.NoopLogger{}, env).RespondWithError(rec, cause, nil)

		s.Equal(http.StatusInternalServerError, rec.Code)

		var out dto.ErrorOutput
		s.Require().NoError(json.NewDecoder(rec.Body).Decode(&out))

		return out
	}

	s.Run("should expose the cause in development and test", func() {
		for _, env := range []string{config.EnvironmentDevelopment, config.EnvironmentTest} {
			s.Equal(cause.Error(), respond(env).Cause, env)
		}
	})

	s.Run("should hide the cause in production, empty and unknown environments", func() {
		for _, env := range []string{config.EnvironmentProduction, "", "staging", "Development"} {
			s.Empty(respond(env).Cause, env)
		}
	})
}

func (s *ResponseTestSuite) TestKindToStatus() {
	s.Run("should map the payload too large kind to 413", func() {
		s.Equal(http.StatusRequestEntityTooLarge, kindToStatus(apperror.PayloadTooLargeKind))
	})
}
