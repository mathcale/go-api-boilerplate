package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID      `db:"id"`
	Name      string         `db:"name"`
	Surname   string         `db:"surname"`
	AvatarURL *string        `db:"avatar_url"`
	Email     string         `db:"email"`
	Password  string         `db:"password"`
	Active    bool           `db:"active"`
	Roles     pq.StringArray `db:"roles"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}
