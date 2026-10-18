package dto

import "time"

type (
	SignUpInput struct {
		Name      string  `json:"name" validate:"required,max=255"`
		Surname   string  `json:"surname" validate:"required,max=255"`
		AvatarURL *string `json:"avatar_url" validate:"omitempty,http_url,max=2048"`
		Email     string  `json:"email" validate:"required,email,max=255"`
		Password  string  `json:"password" validate:"required,min=8,maxbytes=72"`
	}

	SignUpOutput struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Surname string `json:"surname"`
		Email   string `json:"email"`
		Active  bool   `json:"active"`
	}

	SignInInput struct {
		Email    string `json:"email" validate:"required,email,max=255"`
		Password string `json:"password" validate:"required,maxbytes=72"`
	}

	TokenPairOutput struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	RefreshTokenInput struct {
		RefreshToken string `json:"refresh_token" validate:"required,max=4096"`
	}

	ConfirmAccountInput struct {
		UserID string `json:"user_id" validate:"required,uuid,max=64"`
		Code   string `json:"code" validate:"required,max=64"`
	}

	ResendConfirmationCodeInput struct {
		Email string `json:"email" validate:"required,email,max=255"`
	}

	SetRecoveryCodeInput struct {
		Email string `json:"email" validate:"required,email,max=255"`
	}

	ValidateRecoveryCodeInput struct {
		UserID string `json:"user_id" validate:"required,uuid,max=64"`
		Code   string `json:"code" validate:"required,max=64"`
	}

	UpdatePasswordInput struct {
		UserID      string `json:"user_id" validate:"required,uuid,max=64"`
		Code        string `json:"code" validate:"required,max=64"`
		NewPassword string `json:"new_password" validate:"required,min=8,maxbytes=72"`
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
