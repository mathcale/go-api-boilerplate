package dto

import "github.com/mathcale/go-api-boilerplate/internal/domain"

type SignInInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignInOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewSignInOutput(accessToken, refreshToken string) SignInOutput {
	return SignInOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func (in *SignInInput) Validate() error {
	return validate.Struct(in)
}

func (in *SignInInput) ToDomain() domain.User {
	return domain.User{
		Email:    in.Email,
		Password: in.Password,
	}
}
