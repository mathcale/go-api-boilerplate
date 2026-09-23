package user_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
)

type RefreshTokenDomainTestSuite struct {
	suite.Suite
}

func TestRefreshTokenDomain(t *testing.T) {
	suite.Run(t, new(RefreshTokenDomainTestSuite))
}

func (s *RefreshTokenDomainTestSuite) TestNewRefreshTokenFamily() {
	s.Run("should create a new refresh token that starts its own family", func() {
		id := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().Add(time.Hour)

		rt := user.NewRefreshTokenFamily(id, userID, expiresAt)

		s.Equal(id, rt.ID)
		s.Equal(id, rt.FamilyID)
		s.Equal(userID, rt.UserID)
		s.Equal(expiresAt, rt.ExpiresAt)
		s.Nil(rt.UsedAt)
		s.Nil(rt.RevokedAt)
	})
}

func (s *RefreshTokenDomainTestSuite) TestRotateRefreshToken() {
	s.Run("should rotate a refresh token while keeping the same family and user", func() {
		previous := user.NewRefreshTokenFamily(uuid.New(), uuid.New(), time.Now().Add(time.Hour))
		newID := uuid.New()
		newExpiresAt := time.Now().Add(2 * time.Hour)

		rotated := user.RotateRefreshToken(previous, newID, newExpiresAt)

		s.Equal(newID, rotated.ID)
		s.Equal(previous.FamilyID, rotated.FamilyID)
		s.Equal(previous.UserID, rotated.UserID)
		s.Equal(newExpiresAt, rotated.ExpiresAt)
	})
}

func (s *RefreshTokenDomainTestSuite) TestIsExpired() {
	s.Run("should report whether a refresh token is expired", func() {
		expired := user.RefreshToken{
			ExpiresAt: time.Now().Add(-time.Minute),
		}

		notExpired := user.RefreshToken{
			ExpiresAt: time.Now().Add(time.Minute),
		}

		s.True(expired.IsExpired())
		s.False(notExpired.IsExpired())
	})
}

func (s *RefreshTokenDomainTestSuite) TestIsUsed() {
	s.Run("should report whether a refresh token has been used", func() {
		usedAt := time.Now()

		used := user.RefreshToken{
			UsedAt: &usedAt,
		}

		notUsed := user.RefreshToken{}

		s.True(used.IsUsed())
		s.False(notUsed.IsUsed())
	})
}

func (s *RefreshTokenDomainTestSuite) TestIsRevoked() {
	s.Run("should report whether a refresh token has been revoked", func() {
		revokedAt := time.Now()

		revoked := user.RefreshToken{
			RevokedAt: &revokedAt,
		}

		notRevoked := user.RefreshToken{}

		s.True(revoked.IsRevoked())
		s.False(notRevoked.IsRevoked())
	})
}
