package mappers

import (
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
)

func RefreshTokenModelToDomain(m models.RefreshToken) user.RefreshToken {
	return user.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		FamilyID:  m.FamilyID,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
}

func RefreshTokenDomainToModel(t user.RefreshToken) models.RefreshToken {
	return models.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		FamilyID:  t.FamilyID,
		ExpiresAt: t.ExpiresAt,
		UsedAt:    t.UsedAt,
		RevokedAt: t.RevokedAt,
		CreatedAt: t.CreatedAt,
	}
}
