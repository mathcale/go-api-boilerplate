package gateways

import (
	"context"

	"github.com/mathcale/go-api-boilerplate/internal/domain"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/repositories"
	authuc "github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type signInGateway struct {
	repo repositories.User
}

func NewSignInGateway(repo repositories.User) authuc.SignInGateway {
	return &signInGateway{
		repo: repo,
	}
}

func (gw *signInGateway) GetUser(ctx context.Context, email string) (*domain.User, error) {
	model, err := gw.repo.Get(ctx, email)
	if err != nil {
		return nil, err
	}

	if model == nil {
		return nil, nil
	}

	user := domain.NewUser(
		model.ID,
		model.Name,
		model.Email,
		model.Password,
		model.Active,
		model.CreatedAt,
		model.UpdatedAt,
	)

	return &user, nil
}
