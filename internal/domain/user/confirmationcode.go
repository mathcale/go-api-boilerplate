package user

import (
	"time"

	"github.com/google/uuid"
)

type Purpose string

const (
	PurposeAccountConfirmation Purpose = "account_confirmation"
	PurposePasswordRecovery    Purpose = "password_recovery"

	AccountConfirmationTTL = 10 * time.Minute
	PasswordRecoveryTTL    = 30 * time.Minute
)

type ConfirmationCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Code      string
	Purpose   Purpose
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

func NewConfirmationCode(userID uuid.UUID, purpose Purpose, ttl time.Duration) ConfirmationCode {
	now := time.Now()

	return ConfirmationCode{
		ID:        uuid.New(),
		UserID:    userID,
		Code:      uuid.NewString(),
		Purpose:   purpose,
		ExpiresAt: now.Add(ttl),
		Used:      false,
		CreatedAt: now,
	}
}

func (c *ConfirmationCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}
