package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/gateway"
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type (
	MeUseCase interface {
		Execute(ctx context.Context, userID uuid.UUID) (*user.User, error)
	}

	meUseCase struct {
		logger  logger.Logger
		gateway gateway.User
	}
)

func NewMeUseCase(l logger.Logger, gw gateway.User) MeUseCase {
	return &meUseCase{
		logger:  l,
		gateway: gw,
	}
}

func (uc *meUseCase) Execute(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	target, err := uc.gateway.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if target == nil {
		notFound := apperror.BE1001_USER_NOT_FOUND

		return nil, apperror.New(
			nil, "user not found", apperror.NotFoundKind,
			apperror.UseCaseOrigin, "me", &notFound, nil,
		)
	}

	return target, nil
}
