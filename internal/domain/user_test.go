package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
)

type UserDomainTestSuite struct {
	suite.Suite
	domain domain.User
}

func (s *UserDomainTestSuite) SetupTest() {
	now := time.Now()

	s.domain = domain.User{
		ID:        uuid.New(),
		Name:      "any-name",
		Email:     "foo@example.com",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestUserDomain(t *testing.T) {
	suite.Run(t, new(UserDomainTestSuite))
}

func (s *UserDomainTestSuite) TestNewUser() {
	s.Run("should return new User domain instance", func() {
		rawID := "a48e7460-18d7-4ca2-a470-3cb057c4d5cb"
		id, _ := uuid.Parse(rawID)
		domain := domain.NewUser(
			id,
			"any-name",
			"foo@example.com",
			"any-password",
			true,
			time.Now(),
			time.Now(),
		)

		s.Equal(rawID, domain.ID.String())
		s.Equal("any-name", domain.Name)
		s.Equal("foo@example.com", domain.Email)
		s.Equal("any-password", domain.Password)
		s.True(domain.Active)
	})
}

func (s *UserDomainTestSuite) TestToModel() {
	s.Run("should return new User model instance", func() {
		model := s.domain.ToPartialModel()

		s.Equal(s.domain.Email, model.Email)
		s.Equal(s.domain.Name, model.Name)
		s.Equal(s.domain.Active, model.Active)
	})
}
