package mappers

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/tests/fixtures"
)

type MappersUserTestSuite struct {
	suite.Suite
}

func TestMappersUser(t *testing.T) {
	suite.Run(t, new(MappersUserTestSuite))
}

func (s *MappersUserTestSuite) TestUserModelToDomain() {
	s.Run("should map a user model to its domain representation", func() {
		m := fixtures.DomainPkgUserModel()

		got := UserModelToDomain(m)

		s.Equal(m.ID, got.ID)
		s.Equal(m.Name, got.Name)
		s.Equal(m.Surname, got.Surname)
		s.Equal(m.AvatarURL, got.AvatarURL)
		s.Equal(m.Email, got.Email)
		s.Equal(m.Password, got.Password)
		s.Equal(m.Active, got.Active)
		s.Equal([]string{"admin"}, got.Roles)
		s.Equal(m.CreatedAt, got.CreatedAt)
		s.Equal(m.UpdatedAt, got.UpdatedAt)
		s.Equal(m.TokenVersion, got.TokenVersion)
	})
}

func (s *MappersUserTestSuite) TestUserAuthStateModelToDomain() {
	s.Run("should map an auth state model to its domain representation", func() {
		m := fixtures.DomainPkgUserAuthStateModel()

		got := UserAuthStateModelToDomain(m)

		s.Equal(m.TokenVersion, got.TokenVersion)
		s.Equal(m.Active, got.Active)
	})
}

func (s *MappersUserTestSuite) TestUserDomainToModel() {
	s.Run("should map a domain user to its persistence model", func() {
		u := fixtures.DomainPkgUser()

		got := UserDomainToModel(u)

		s.Equal(u.ID, got.ID)
		s.Equal(u.Name, got.Name)
		s.Equal(u.Surname, got.Surname)
		s.Equal(u.AvatarURL, got.AvatarURL)
		s.Equal(u.Email, got.Email)
		s.Equal(u.Password, got.Password)
		s.Equal(u.Active, got.Active)
		s.Equal(pq.StringArray{"admin"}, got.Roles)
		s.Equal(u.CreatedAt, got.CreatedAt)
		s.Equal(u.UpdatedAt, got.UpdatedAt)
		s.Equal(u.TokenVersion, got.TokenVersion)
	})

	s.Run("should map nil roles to an empty slice instead of null", func() {
		u := user.User{Roles: nil}

		got := UserDomainToModel(u)

		s.Equal(pq.StringArray{}, got.Roles)
	})
}
