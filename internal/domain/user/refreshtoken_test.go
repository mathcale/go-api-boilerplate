package user_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
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
		previous := fixtures.DomainPkgFamilyRefreshToken()
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
	s.Run("should report a past refresh token as expired", func() {
		rt := fixtures.DomainPkgExpiredRefreshToken()

		s.True(rt.IsExpired())
	})

	s.Run("should report a future refresh token as not expired", func() {
		rt := fixtures.DomainPkgActiveRefreshToken()

		s.False(rt.IsExpired())
	})
}

func (s *RefreshTokenDomainTestSuite) TestIsUsed() {
	s.Run("should report a refresh token with a used timestamp as used", func() {
		rt := fixtures.DomainPkgUsedRefreshToken()

		s.True(rt.IsUsed())
	})

	s.Run("should report a refresh token without a used timestamp as not used", func() {
		rt := user.RefreshToken{}

		s.False(rt.IsUsed())
	})
}

func (s *RefreshTokenDomainTestSuite) TestIsRevoked() {
	s.Run("should report a refresh token with a revoked timestamp as revoked", func() {
		rt := fixtures.DomainPkgRevokedRefreshToken()

		s.True(rt.IsRevoked())
	})

	s.Run("should report a refresh token without a revoked timestamp as not revoked", func() {
		rt := user.RefreshToken{}

		s.False(rt.IsRevoked())
	})
}
