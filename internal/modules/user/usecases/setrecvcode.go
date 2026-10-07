package usecases

import (
	"context"
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
)

type (
	SetRecoveryCodeUseCase interface {
		Execute(ctx context.Context, email string) error
	}

	setRecoveryCodeUseCase struct {
		logger      logger.Logger
		gateway     gateway.Gateway
		recoveryTTL time.Duration
	}
)

func NewSetRecoveryCodeUseCase(
	l logger.Logger,
	gw gateway.Gateway,
	recoveryTTL time.Duration,
) SetRecoveryCodeUseCase {
	return &setRecoveryCodeUseCase{
		logger:      l,
		gateway:     gw,
		recoveryTTL: recoveryTTL,
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
		uc.recoveryTTL,
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
