package user_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/tests/fixtures"
)

type UserDomainTestSuite struct {
	suite.Suite
}

func TestUserDomain(t *testing.T) {
	suite.Run(t, new(UserDomainTestSuite))
}

func (s *UserDomainTestSuite) TestCreateFromInput() {
	s.Run("should create a user trimming whitespace and lowercasing the email", func() {
		created, err := user.CreateFromInput(fixtures.DomainPkgCreateUserParams())

		s.Require().NoError(err)
		s.NotEmpty(created.ID)
		s.Equal("Jane", created.Name)
		s.Equal("Doe", created.Surname)
		s.Equal("jane@example.com", created.Email)
		s.False(created.Active)
		s.WithinDuration(created.CreatedAt, created.UpdatedAt, 0)
	})

	s.Run("should return an error when name is missing", func() {
		params := fixtures.DomainPkgCreateUserParams()
		params.Name = "   "

		_, err := user.CreateFromInput(params)

		s.Require().EqualError(err, "name is required")
	})

	s.Run("should return an error when surname is missing", func() {
		params := fixtures.DomainPkgCreateUserParams()
		params.Surname = ""

		_, err := user.CreateFromInput(params)

		s.Require().EqualError(err, "surname is required")
	})

	s.Run("should return an error when email is missing", func() {
		params := fixtures.DomainPkgCreateUserParams()
		params.Email = ""

		_, err := user.CreateFromInput(params)

		s.Require().EqualError(err, "email is required")
	})

	s.Run("should return an error when password is missing", func() {
		params := fixtures.DomainPkgCreateUserParams()
		params.Password = ""

		_, err := user.CreateFromInput(params)

		s.Require().EqualError(err, "password is required")
	})
}

func (s *UserDomainTestSuite) TestNormalizeEmail() {
	s.Run("should trim whitespace and lowercase the email", func() {
		s.Equal("jane@example.com", user.NormalizeEmail("  Jane@Example.COM \n"))
	})
}

func (s *UserDomainTestSuite) TestValidatePassword() {
	s.Run("should accept a password within the length bounds", func() {
		s.Require().NoError(user.ValidatePassword("secret123"))
	})

	s.Run("should accept a password of exactly the maximum byte length", func() {
		s.Require().NoError(user.ValidatePassword(strings.Repeat("a", user.MaxPasswordBytes)))
	})

	s.Run("should reject a blank password", func() {
		s.Require().EqualError(user.ValidatePassword("   "), "password is required")
	})

	s.Run("should reject a password shorter than the minimum", func() {
		s.Require().EqualError(user.ValidatePassword("short"), "password is too short")
	})

	s.Run("should reject a password longer than the maximum byte length", func() {
		s.Require().EqualError(
			user.ValidatePassword(strings.Repeat("a", user.MaxPasswordBytes+1)),
			"password is too long",
		)
	})

	s.Run("should count bytes rather than runes", func() {
		s.Require().EqualError(
			user.ValidatePassword(strings.Repeat("é", user.MaxPasswordBytes/2+1)),
			"password is too long",
		)
	})
}
