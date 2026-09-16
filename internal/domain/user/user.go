package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Surname   string
	AvatarURL *string
	Email     string
	Password  string
	Active    bool
	Roles     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUserParams struct {
	Name      string
	Surname   string
	AvatarURL *string
	Email     string
	Password  string
	Active    bool
	Roles     []string
}

func CreateFromInput(params CreateUserParams) (User, error) {
	if err := validateUserData(params); err != nil {
		return User{}, err
	}

	now := time.Now()

	return User{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(params.Name),
		Surname:   strings.TrimSpace(params.Surname),
		AvatarURL: params.AvatarURL,
		Email:     strings.ToLower(strings.TrimSpace(params.Email)),
		Password:  params.Password,
		Active:    params.Active,
		Roles:     params.Roles,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func validateUserData(params CreateUserParams) error {
	if strings.TrimSpace(params.Name) == "" {
		return errors.New("name is required")
	}

	if strings.TrimSpace(params.Surname) == "" {
		return errors.New("surname is required")
	}

	if strings.TrimSpace(params.Email) == "" {
		return errors.New("email is required")
	}

	if strings.TrimSpace(params.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}
