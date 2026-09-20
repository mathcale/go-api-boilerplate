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

type MappersRefreshTokenTestSuite struct {
	suite.Suite
}

func TestMappersRefreshToken(t *testing.T) {
	suite.Run(t, new(MappersRefreshTokenTestSuite))
}

func (s *MappersRefreshTokenTestSuite) TestRefreshTokenModelToDomain() {
	s.Run("should map a refresh token model to its domain representation", func() {
		now := time.Now()
		usedAt := now
		revokedAt := now

		m := models.RefreshToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			FamilyID:  uuid.New(),
			ExpiresAt: now,
			UsedAt:    &usedAt,
			RevokedAt: &revokedAt,
			CreatedAt: now,
		}

		got := mappers.RefreshTokenModelToDomain(m)

		s.Equal(m.ID, got.ID)
		s.Equal(m.UserID, got.UserID)
		s.Equal(m.FamilyID, got.FamilyID)
		s.Equal(m.ExpiresAt, got.ExpiresAt)
		s.Equal(m.UsedAt, got.UsedAt)
		s.Equal(m.RevokedAt, got.RevokedAt)
		s.Equal(m.CreatedAt, got.CreatedAt)
	})
}

func (s *MappersRefreshTokenTestSuite) TestRefreshTokenDomainToModel() {
	s.Run("should map a domain refresh token to its persistence model", func() {
		now := time.Now()

		rt := user.RefreshToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			FamilyID:  uuid.New(),
			ExpiresAt: now,
			UsedAt:    nil,
			RevokedAt: nil,
			CreatedAt: now,
		}

		got := mappers.RefreshTokenDomainToModel(rt)

		s.Equal(rt.ID, got.ID)
		s.Equal(rt.UserID, got.UserID)
		s.Equal(rt.FamilyID, got.FamilyID)
		s.Equal(rt.ExpiresAt, got.ExpiresAt)
		s.Nil(got.UsedAt)
		s.Nil(got.RevokedAt)
		s.Equal(rt.CreatedAt, got.CreatedAt)
	})
}
