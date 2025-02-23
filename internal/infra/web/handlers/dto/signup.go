package dto

import (
	"github.com/mathcale/go-api-boilerplate/internal/domain"
)

type SignUpInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func (in *SignUpInput) Validate() error {
	return validate.Struct(in)
}

func (in *SignUpInput) ToDomain(active bool) domain.User {
	return domain.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: in.Password,
		Active:   active,
	}
}
