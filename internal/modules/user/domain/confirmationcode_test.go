package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
)

type ConfirmationCodeTestSuite struct {
	suite.Suite
}

func TestConfirmationCode(t *testing.T) {
	suite.Run(t, new(ConfirmationCodeTestSuite))
}

func (s *ConfirmationCodeTestSuite) TestNewConfirmationCode() {
	s.Run("should create a new confirmation code with the given purpose and expiry", func() {
		userID := uuid.New()

		code := domain.NewConfirmationCode(userID, domain.PurposeAccountConfirmation, domain.AccountConfirmationTTL)

		s.NotEmpty(code.ID)
		s.Equal(userID, code.UserID)
		s.NotEmpty(code.Code)
		s.Equal(domain.PurposeAccountConfirmation, code.Purpose)
		s.False(code.Used)
		s.WithinDuration(time.Now().Add(domain.AccountConfirmationTTL), code.ExpiresAt, time.Second)
	})
}

func (s *ConfirmationCodeTestSuite) TestIsExpired() {
	s.Run("should report whether a confirmation code is expired", func() {
		expired := domain.ConfirmationCode{
			ExpiresAt: time.Now().Add(-time.Minute),
		}

		notExpired := domain.ConfirmationCode{
			ExpiresAt: time.Now().Add(time.Minute),
		}

		s.True(expired.IsExpired())
		s.False(notExpired.IsExpired())
	})
}
