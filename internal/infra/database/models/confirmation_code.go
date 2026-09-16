package models

import (
	"time"

	"github.com/google/uuid"
)

type ConfirmationCode struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Code      string    `db:"code"`
	Purpose   string    `db:"purpose"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
}
