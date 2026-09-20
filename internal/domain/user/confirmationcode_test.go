package user_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
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

		code := user.NewConfirmationCode(userID, user.PurposeAccountConfirmation, user.AccountConfirmationTTL)

		s.NotEmpty(code.ID)
		s.Equal(userID, code.UserID)
		s.NotEmpty(code.Code)
		s.Equal(user.PurposeAccountConfirmation, code.Purpose)
		s.False(code.Used)
		s.WithinDuration(time.Now().Add(user.AccountConfirmationTTL), code.ExpiresAt, time.Second)
	})
}

func (s *ConfirmationCodeTestSuite) TestConfirmationCode_IsExpired() {
	s.Run("should report whether a confirmation code is expired", func() {
		expired := user.ConfirmationCode{
			ExpiresAt: time.Now().Add(-time.Minute),
		}

		notExpired := user.ConfirmationCode{
			ExpiresAt: time.Now().Add(time.Minute),
		}

		s.True(expired.IsExpired())
		s.False(notExpired.IsExpired())
	})
}
