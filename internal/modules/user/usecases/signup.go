package usecases

import (
	"context"
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/core/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
)

type (
	SignUpUseCase interface {
		Execute(ctx context.Context, input SignUpInput) (*domain.User, error)
	}

	SignUpInput struct {
		Name      string
		Surname   string
		AvatarURL *string
		Email     string
		Password  string
	}

	signUpUseCase struct {
		logger          logger.Logger
		gateway         gateway.Gateway
		password        bcrypt.Password
		confirmationTTL time.Duration
	}
)

func NewSignUpUseCase(
	l logger.Logger,
	gw gateway.Gateway,
	pw bcrypt.Password,
	confirmationTTL time.Duration,
) SignUpUseCase {
	return &signUpUseCase{
		logger:          l,
		gateway:         gw,
		password:        pw,
		confirmationTTL: confirmationTTL,
	}
}

func (uc *signUpUseCase) Execute(ctx context.Context, input SignUpInput) (*domain.User, error) {
	exists, err := uc.gateway.UserExistsIncludingInactive(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if exists != nil && *exists {
		code := apperror.BE1000_USER_ALREADY_EXISTS

		return nil, apperror.New(
			nil, "user already exists", apperror.ConflictKind,
			apperror.UseCaseOrigin, "signup", &code, nil,
		)
	}

	hashed, err := uc.password.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	newUser, err := domain.CreateFromInput(domain.CreateUserParams{
		Name:      input.Name,
		Surname:   input.Surname,
		AvatarURL: input.AvatarURL,
		Email:     input.Email,
		Password:  *hashed,
		Active:    false,
	})
	if err != nil {
		code := apperror.BE0001_INVALID_INPUT

		return nil, apperror.New(
			err, "invalid sign-up input", apperror.ValidationKind,
			apperror.UseCaseOrigin, "signup", &code, nil,
		)
	}

	confirmationCode := domain.NewConfirmationCode(
		newUser.ID,
		domain.PurposeAccountConfirmation,
		uc.confirmationTTL,
	)

	if err := uc.gateway.SaveUser(ctx, newUser, confirmationCode); err != nil {
		return nil, err
	}

	if err := uc.gateway.SendConfirmationEmail(ctx, newUser, confirmationCode); err != nil {
		uc.logger.Error("failed to send confirmation email", err, map[string]interface{}{
			logFieldUserID: newUser.ID.String(),
		})
	}

	return &newUser, nil
}
