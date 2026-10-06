package dto

import "time"

type (
	SignUpInput struct {
		Name      string  `json:"name" validate:"required"`
		Surname   string  `json:"surname" validate:"required"`
		AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
		Email     string  `json:"email" validate:"required,email"`
		Password  string  `json:"password" validate:"required,min=8"`
	}

	SignUpOutput struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Surname string `json:"surname"`
		Email   string `json:"email"`
		Active  bool   `json:"active"`
	}

	SignInInput struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	TokenPairOutput struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	RefreshTokenInput struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	ConfirmAccountInput struct {
		UserID string `json:"user_id" validate:"required,uuid"`
		Code   string `json:"code" validate:"required"`
	}

	ResendConfirmationCodeInput struct {
		Email string `json:"email" validate:"required,email"`
	}

	SetRecoveryCodeInput struct {
		Email string `json:"email" validate:"required,email"`
	}

	ValidateRecoveryCodeInput struct {
		UserID string `json:"user_id" validate:"required,uuid"`
		Code   string `json:"code" validate:"required"`
	}

	UpdatePasswordInput struct {
		UserID      string `json:"user_id" validate:"required,uuid"`
		Code        string `json:"code" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	MeOutput struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Surname   string    `json:"surname"`
		AvatarURL *string   `json:"avatar_url"`
		Email     string    `json:"email"`
		Active    bool      `json:"active"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)
