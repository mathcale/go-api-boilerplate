package auth

import (
	"context"
	"errors"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type SignUpUseCase interface {
	Execute(ctx context.Context, user domain.User) error
}

type SignUpGateway interface {
	UserExists(ctx context.Context, email string) (*bool, error)
	SaveUser(ctx context.Context, user domain.User) error
}

type signUpUseCase struct {
	logger     logger.Logger
	bcryptPass bcrypt.Password
	gateway    SignUpGateway
}

func NewSignUpUseCase(
	l logger.Logger,
	bcryptPass bcrypt.Password,
	gw SignUpGateway,
) SignUpUseCase {
	return &signUpUseCase{
		logger:     l,
		bcryptPass: bcryptPass,
		gateway:    gw,
	}
}

func (s *signUpUseCase) Execute(ctx context.Context, user domain.User) error {
	s.logger.Debug("Checking if user exists", map[string]interface{}{
		"email": user.Email,
	})

	exists, err := s.gateway.UserExists(ctx, user.Email)
	if err != nil {
		return err
	}

	if *exists {
		return apperror.New(
			errors.New("user_already_exists"), "user already exists",
			apperror.ConflictKind, apperror.UseCaseOrigin, "signup", nil, nil,
		)
	}

	hashedPass, err := s.bcryptPass.Hash(user.Password)
	if err != nil {
		return err
	}

	user.Password = *hashedPass

	s.logger.Debug("Saving new user", map[string]interface{}{
		"email": user.Email,
	})

	return s.gateway.SaveUser(ctx, user)
}
