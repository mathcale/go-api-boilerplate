package fixtures

import (
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
)

var (
	id       = "a48e7460-18d7-4ca2-a470-3cb057c4d5cb"
	password = "any-password"
)

func NewUserDomain() domain.User {
	return domain.User{
		Name:     "any-name",
		Email:    "foo@example.com",
		Password: password,
		Active:   true,
	}
}

func NewUserModel() models.User {
	return models.NewUser("any-name", "foo@example.com", password, true)
}

func NewUserModelComplete() models.User {
	now := time.Now()

	return models.User{
		ID:        generateUUID(),
		Name:      "any-name",
		Email:     "foo@example.com",
		Password:  hashPassword(),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewUserDomainComplete() domain.User {
	now := time.Now()

	return domain.User{
		ID:        generateUUID(),
		Name:      "any-name",
		Email:     "foo@example.com",
		Password:  hashPassword(),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func generateUUID() uuid.UUID {
	id, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Error while generating UUID from static value: %v", err)
	}

	return id
}

func hashPassword() string {
	hashedPasswd, _ := bcrypt.NewPassword().Hash(password)
	return *hashedPasswd
}
