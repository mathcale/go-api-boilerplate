package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/infra/web/handlers/dto"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/tests/fixtures"
)

type RequestTestSuite struct {
	suite.Suite
}

func TestParseUUID(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}

func (s *RequestTestSuite) TestParseUUID() {
	s.Run("should return the parsed uuid when the string is a valid uuid", func() {
		id, err := parseUUID("123e4567-e89b-12d3-a456-426614174000", "test_origin")

		s.Require().NoError(err)
		s.Equal("123e4567-e89b-12d3-a456-426614174000", id.String())
	})

	s.Run("should return an error when the string is not a valid uuid", func() {
		_, err := parseUUID("not-a-uuid", "test_origin")

		s.Require().Error(err)

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
	})
}

func (s *RequestTestSuite) TestDecodeAndValidate() {
	type payload struct {
		Name string `json:"name" validate:"required"`
	}

	newRequest := func(body string) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	}

	s.Run("should decode the body into dst when the payload is valid", func() {
		var dst payload

		err := decodeAndValidate(newRequest(`{"name":"john"}`), &dst, "test_origin")

		s.Require().NoError(err)
		s.Equal("john", dst.Name)
	})

	s.Run("should return a parse error when the body is not valid json", func() {
		var dst payload

		err := decodeAndValidate(newRequest(`{invalid`), &dst, "test_origin")

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ParseKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
	})

	s.Run("should return a payload too large error when the body exceeds the limit", func() {
		var dst struct {
			Email string `json:"email"`
		}

		req := newRequest(fixtures.WebOversizedBody(2048))
		req.Body = http.MaxBytesReader(httptest.NewRecorder(), req.Body, 1024)

		err := decodeAndValidate(req, &dst, "test_origin")

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.PayloadTooLargeKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0006_PAYLOAD_TOO_LARGE, *appErr.BusinessCode())
	})

	s.Run("should return a parse error when the body has an unknown field", func() {
		var dst dto.SignInInput

		err := decodeAndValidate(newRequest(fixtures.WebBodyWithUnknownField()), &dst, "o")

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ParseKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
	})

	s.Run("should return a parse error when data trails the json object", func() {
		for _, body := range []string{
			fixtures.WebBodyWithTrailingData(), fixtures.WebBodyWithTrailingGarbage(),
		} {
			var dst dto.SignInInput

			err := decodeAndValidate(newRequest(body), &dst, "test_origin")

			var appErr apperror.AppError
			s.Require().ErrorAs(err, &appErr)
			s.Equal(apperror.ParseKind, appErr.Kind())
			s.Require().NotNil(appErr.BusinessCode())
			s.Equal(apperror.BE0001_INVALID_INPUT, *appErr.BusinessCode())
		}
	})

	s.Run("should return a validation error when the payload fails validation", func() {
		var dst payload

		err := decodeAndValidate(newRequest(`{}`), &dst, "test_origin")

		var appErr apperror.AppError
		s.Require().ErrorAs(err, &appErr)
		s.Equal(apperror.ValidationKind, appErr.Kind())
		s.Require().NotNil(appErr.BusinessCode())
		s.Equal(apperror.BE0002_INPUT_VALIDATION_FAILED, *appErr.BusinessCode())
	})
}
