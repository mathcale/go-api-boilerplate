package user

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinPasswordRunes = 8
	MaxPasswordBytes = 72
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

	TokenVersion int
}

type AuthState struct {
	TokenVersion int
	Active       bool
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
		Email:     NormalizeEmail(params.Email),
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

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}

	if utf8.RuneCountInString(password) < MinPasswordRunes {
		return errors.New("password is too short")
	}

	if len(password) > MaxPasswordBytes {
		return errors.New("password is too long")
	}

	return nil
}
