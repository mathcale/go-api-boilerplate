package user_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
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

		code := user.NewConfirmationCode(
			userID, user.PurposeAccountConfirmation, user.AccountConfirmationTTL,
		)

		s.NotEmpty(code.ID)
		s.Equal(userID, code.UserID)
		s.NotEmpty(code.Code)
		s.Equal(user.PurposeAccountConfirmation, code.Purpose)
		s.False(code.Used)
		s.WithinDuration(time.Now().Add(user.AccountConfirmationTTL), code.ExpiresAt, time.Second)
	})
}

func (s *ConfirmationCodeTestSuite) TestIsExpired() {
	s.Run("should report a past confirmation code as expired", func() {
		code := fixtures.DomainPkgExpiredConfirmationCode()

		s.True(code.IsExpired())
	})

	s.Run("should report a future confirmation code as not expired", func() {
		code := fixtures.DomainPkgActiveConfirmationCode()

		s.False(code.IsExpired())
	})
}
