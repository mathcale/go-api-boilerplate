package mappers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/tests/fixtures"
)

type MappersConfirmationCodeTestSuite struct {
	suite.Suite
}

func TestMappersConfirmationCode(t *testing.T) {
	suite.Run(t, new(MappersConfirmationCodeTestSuite))
}

func (s *MappersConfirmationCodeTestSuite) TestConfirmationCodeModelToDomain() {
	s.Run("should map a confirmation code model to its domain representation", func() {
		m := fixtures.DomainPkgConfirmationCodeModel()

		got := ConfirmationCodeModelToDomain(m)

		s.Equal(m.ID, got.ID)
		s.Equal(m.UserID, got.UserID)
		s.Equal(m.Code, got.Code)
		s.Equal(user.Purpose(m.Purpose), got.Purpose)
		s.Equal(m.ExpiresAt, got.ExpiresAt)
		s.Equal(m.Used, got.Used)
		s.Equal(m.CreatedAt, got.CreatedAt)
	})
}

func (s *MappersConfirmationCodeTestSuite) TestConfirmationCodeDomainToModel() {
	s.Run("should map a domain confirmation code to its persistence model", func() {
		c := fixtures.DomainPkgConfirmationCode()

		got := ConfirmationCodeDomainToModel(c)

		s.Equal(c.ID, got.ID)
		s.Equal(c.UserID, got.UserID)
		s.Equal(c.Code, got.Code)
		s.Equal(string(c.Purpose), got.Purpose)
		s.Equal(c.ExpiresAt, got.ExpiresAt)
		s.Equal(c.Used, got.Used)
		s.Equal(c.CreatedAt, got.CreatedAt)
	})
}
