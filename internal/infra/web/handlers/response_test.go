package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/infra/web/handlers/dto"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/pkg/logger"
	"go-api-boilerplate/internal/tests/mocks"
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

	s.Run("should log the app error metadata for server errors", func() {
		buf := &bytes.Buffer{}
		code := apperror.BE0000_UNKNOWN_ERROR
		appErr := apperror.New(
			cause, "find user", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "get_by_id", &code, nil,
		)

		rec := httptest.NewRecorder()
		NewResponse(logger.NewWithWriter("info", buf), config.EnvironmentProduction).
			RespondWithError(rec, appErr, nil)

		s.Equal(http.StatusInternalServerError, rec.Code)

		entry := map[string]any{}
		s.Require().NoError(json.Unmarshal(buf.Bytes(), &entry))
		s.Equal("find user: db password leaked", entry["error"])
		s.Equal(string(apperror.DatabaseKind), entry["kind"])
		s.Equal(string(apperror.RepositoryOrigin), entry["origin"])
		s.Equal("get_by_id", entry["origin_name"])
		s.Equal(string(code), entry["business_code"])
		s.EqualValues(http.StatusInternalServerError, entry["status"])
	})

	s.Run("should not log client errors", func() {
		buf := &bytes.Buffer{}
		appErr := apperror.New(
			cause, "user not found", apperror.NotFoundKind,
			apperror.UseCaseOrigin, "me", nil, nil,
		)

		rec := httptest.NewRecorder()
		NewResponse(logger.NewWithWriter("info", buf), config.EnvironmentProduction).
			RespondWithError(rec, appErr, nil)

		s.Equal(http.StatusNotFound, rec.Code)
		s.Empty(buf.String())
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
