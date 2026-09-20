package mappers_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/mappers"
)

type MappersConfirmationCodeTestSuite struct {
	suite.Suite
}

func TestMappersConfirmationCode(t *testing.T) {
	suite.Run(t, new(MappersConfirmationCodeTestSuite))
}

func (s *MappersConfirmationCodeTestSuite) TestConfirmationCodeModelToDomain() {
	s.Run("should map a confirmation code model to its domain representation", func() {
		now := time.Now()

		m := models.ConfirmationCode{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Code:      "the-code",
			Purpose:   "account_confirmation",
			ExpiresAt: now,
			Used:      true,
			CreatedAt: now,
		}

		got := mappers.ConfirmationCodeModelToDomain(m)

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
		now := time.Now()

		c := user.ConfirmationCode{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Code:      "the-code",
			Purpose:   user.PurposePasswordRecovery,
			ExpiresAt: now,
			Used:      false,
			CreatedAt: now,
		}

		got := mappers.ConfirmationCodeDomainToModel(c)

		s.Equal(c.ID, got.ID)
		s.Equal(c.UserID, got.UserID)
		s.Equal(c.Code, got.Code)
		s.Equal(string(c.Purpose), got.Purpose)
		s.Equal(c.ExpiresAt, got.ExpiresAt)
		s.Equal(c.Used, got.Used)
		s.Equal(c.CreatedAt, got.CreatedAt)
	})
}
