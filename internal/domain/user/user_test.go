package user_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
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
