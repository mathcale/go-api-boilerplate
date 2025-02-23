package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Password  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(
	id uuid.UUID, name string, email string, password string,
	active bool, createdAt time.Time, updatedAt time.Time,
) User {
	return User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  password,
		Active:    active,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (u *User) ToPartialModel() models.User {
	return models.NewUser(u.Name, u.Email, u.Password, u.Active)
}
