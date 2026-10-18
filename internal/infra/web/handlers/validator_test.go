package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type RequestValidatorTestSuite struct {
	suite.Suite
}

func TestNewRequestValidator(t *testing.T) {
	suite.Run(t, new(RequestValidatorTestSuite))
}

func (s *RequestValidatorTestSuite) TestStruct() {
	v := newRequestValidator()

	s.Run("should accept a valid sign-up payload with an https avatar", func() {
		s.Require().NoError(v.Struct(fixtures.WebValidSignUpInput()))
	})

	s.Run("should accept a sign-up payload without an avatar", func() {
		in := fixtures.WebValidSignUpInput()
		in.AvatarURL = nil

		s.Require().NoError(v.Struct(in))
	})

	s.Run("should accept a 72 byte password", func() {
		in := fixtures.WebSignUpInputWithPassword(strings.Repeat("a", 72))

		s.Require().NoError(v.Struct(in))
	})

	s.Run("should reject a 73 byte password", func() {
		in := fixtures.WebSignUpInputWithPassword(strings.Repeat("a", 73))

		s.Require().Error(v.Struct(in))
	})

	s.Run("should reject a password within 72 runes but over 72 bytes", func() {
		in := fixtures.WebSignUpInputWithPassword(strings.Repeat("é", 40))

		s.Require().Error(v.Struct(in))
	})

	s.Run("should reject over-long sign-in and update-password passwords", func() {
		long := strings.Repeat("a", 73)

		s.Require().Error(v.Struct(dto.SignInInput{Email: fixtures.WebEmail, Password: long}))
		s.Require().Error(v.Struct(dto.UpdatePasswordInput{
			UserID:      fixtures.WebUserID().String(),
			Code:        fixtures.WebCode,
			NewPassword: long,
		}))
	})

	s.Run("should reject non-http avatar urls", func() {
		for _, avatar := range []string{
			"javascript:alert(1)", "data:text/html;base64,PHNjcmlwdD4=", "ftp://example.com/a.png",
		} {
			s.Require().Error(v.Struct(fixtures.WebSignUpInputWithAvatar(avatar)), avatar)
		}
	})

	s.Run("should reject an avatar url over 2048 characters", func() {
		avatar := "https://example.com/" + strings.Repeat("a", 2048)

		s.Require().Error(v.Struct(fixtures.WebSignUpInputWithAvatar(avatar)))
	})

	s.Run("should reject over-length name and email", func() {
		s.Require().Error(v.Struct(fixtures.WebSignUpInputWithName(strings.Repeat("a", 256))))
		s.Require().Error(v.Struct(
			fixtures.WebSignUpInputWithEmail(strings.Repeat("a", 250) + "@x.com"),
		))
	})

	s.Run("should reject over-length codes and refresh tokens", func() {
		s.Require().Error(v.Struct(dto.ConfirmAccountInput{
			UserID: fixtures.WebUserID().String(),
			Code:   strings.Repeat("a", 65),
		}))
		s.Require().Error(v.Struct(
			dto.RefreshTokenInput{RefreshToken: strings.Repeat("a", 4097)},
		))
	})
}
