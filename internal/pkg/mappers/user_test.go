package mappers_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/mappers"
)

type MappersUserTestSuite struct {
	suite.Suite
}

func TestMappersUser(t *testing.T) {
	suite.Run(t, new(MappersUserTestSuite))
}

func (s *MappersUserTestSuite) TestUserModelToDomain() {
	s.Run("should map a user model to its domain representation", func() {
		avatarURL := "https://example.com/avatar.png"
		now := time.Now()

		m := models.User{
			ID:        uuid.New(),
			Name:      "Jane",
			Surname:   "Doe",
			AvatarURL: &avatarURL,
			Email:     "jane@example.com",
			Password:  "hashed",
			Active:    true,
			Roles:     pq.StringArray{"admin"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		got := mappers.UserModelToDomain(m)

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
	})
}

func (s *MappersUserTestSuite) TestUserDomainToModel() {
	s.Run("should map a domain user to its persistence model", func() {
		avatarURL := "https://example.com/avatar.png"
		now := time.Now()

		u := user.User{
			ID:        uuid.New(),
			Name:      "Jane",
			Surname:   "Doe",
			AvatarURL: &avatarURL,
			Email:     "jane@example.com",
			Password:  "hashed",
			Active:    true,
			Roles:     []string{"admin"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		got := mappers.UserDomainToModel(u)

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
	})

	s.Run("should map nil roles to an empty slice instead of null", func() {
		u := user.User{Roles: nil}

		got := mappers.UserDomainToModel(u)

		s.Equal(pq.StringArray{}, got.Roles)
	})
}
