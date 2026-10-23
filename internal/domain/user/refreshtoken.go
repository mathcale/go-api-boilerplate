package user

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FamilyID  uuid.UUID
	ExpiresAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func NewRefreshTokenFamily(id, userID uuid.UUID, expiresAt time.Time) RefreshToken {
	return RefreshToken{
		ID:        id,
		UserID:    userID,
		FamilyID:  id,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

func RotateRefreshToken(previous RefreshToken, newID uuid.UUID, expiresAt time.Time) RefreshToken {
	return RefreshToken{
		ID:        newID,
		UserID:    previous.UserID,
		FamilyID:  previous.FamilyID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *RefreshToken) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}
