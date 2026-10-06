package usecases

import (
	"context"

	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
)

type (
	SetRecoveryCodeUseCase interface {
		Execute(ctx context.Context, email string) error
	}

	setRecoveryCodeUseCase struct {
		logger  logger.Logger
		gateway gateway.Gateway
	}
)

func NewSetRecoveryCodeUseCase(l logger.Logger, gw gateway.Gateway) SetRecoveryCodeUseCase {
	return &setRecoveryCodeUseCase{
		logger:  l,
		gateway: gw,
	}
}

func (uc *setRecoveryCodeUseCase) Execute(ctx context.Context, email string) error {
	target, err := uc.gateway.GetUserByEmailIncludingInactive(ctx, email)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	recoveryCode := domain.NewConfirmationCode(
		target.ID,
		domain.PurposePasswordRecovery,
		domain.PasswordRecoveryTTL,
	)

	if err := uc.gateway.SaveConfirmationCode(ctx, recoveryCode); err != nil {
		return err
	}

	if err := uc.gateway.SendRecoveryEmail(ctx, *target, recoveryCode); err != nil {
		uc.logger.Error("failed to send recovery email", err, map[string]interface{}{
			logFieldUserID: target.ID.String(),
		})
	}

	return nil
}
