package mappers

import (
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/internal/models"
)

func RefreshTokenModelToDomain(m models.RefreshToken) domain.RefreshToken {
	return domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		FamilyID:  m.FamilyID,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
}

func RefreshTokenDomainToModel(t domain.RefreshToken) models.RefreshToken {
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
