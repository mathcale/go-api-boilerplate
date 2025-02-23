package gateways

import (
	"context"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/repositories"
	authuc "github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type signUpGateway struct {
	repo repositories.User
}

func NewSignUpGateway(repo repositories.User) authuc.SignUpGateway {
	return &signUpGateway{
		repo: repo,
	}
}

func (gw *signUpGateway) UserExists(ctx context.Context, email string) (*bool, error) {
	return gw.repo.Exists(ctx, email)
}

func (gw *signUpGateway) SaveUser(ctx context.Context, user domain.User) error {
	return gw.repo.Save(ctx, user.ToPartialModel())
}
