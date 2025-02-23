package auth

import (
	"context"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type SignInUseCase interface {
	Execute(ctx context.Context, user domain.User) (accessToken, refreshToken *string, err error)
}

type SignInGateway interface {
	GetUser(ctx context.Context, email string) (*domain.User, error)
}

type signInUseCase struct {
	logger  logger.Logger
	passwd  bcrypt.Password
	jwt     jwt.JWTAuth
	gateway SignInGateway
}

func NewSignInUseCase(
	l logger.Logger,
	passwd bcrypt.Password,
	jwt jwt.JWTAuth,
	gw SignInGateway,
) SignInUseCase {
	return &signInUseCase{
		logger:  l,
		passwd:  passwd,
		jwt:     jwt,
		gateway: gw,
	}
}

func (uc *signInUseCase) Execute(ctx context.Context, user domain.User) (*string, *string, error) {
	foundUser, err := uc.gateway.GetUser(ctx, user.Email)
	if err != nil {
		return nil, nil, err
	}

	if err = uc.verifyPassword(user.Password, foundUser.Password); err != nil {
		return nil, nil, err
	}

	at, err := uc.issueAccessToken(foundUser.ID.String())
	if err != nil {
		return nil, nil, err
	}

	rt, err := uc.issueRefreshToken(foundUser.ID.String())
	if err != nil {
		return nil, nil, err
	}

	return at.Token, rt.Token, nil
}

func (uc *signInUseCase) verifyPassword(plain string, hashed string) error {
	return uc.passwd.Verify(plain, hashed)
}

func (uc *signInUseCase) issueAccessToken(id string) (*jwt.Token, error) {
	return uc.jwt.IssueAccessToken(id)
}

func (uc *signInUseCase) issueRefreshToken(id string) (*jwt.Token, error) {
	return uc.jwt.IssueRefreshToken(id)
}
