package usecases

import (
	"context"
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
)

type (
	ResendConfirmationCodeUseCase interface {
		Execute(ctx context.Context, email string) error
	}

	resendConfirmationCodeUseCase struct {
		logger          logger.Logger
		gateway         gateway.Gateway
		confirmationTTL time.Duration
	}
)

func NewResendConfirmationCodeUseCase(
	l logger.Logger,
	gw gateway.Gateway,
	confirmationTTL time.Duration,
) ResendConfirmationCodeUseCase {
	return &resendConfirmationCodeUseCase{
		logger:          l,
		gateway:         gw,
		confirmationTTL: confirmationTTL,
	}
}

func (uc *resendConfirmationCodeUseCase) Execute(ctx context.Context, email string) error {
	target, err := uc.gateway.GetUserByEmailIncludingInactive(ctx, email)
	if err != nil {
		return err
	}

	if target == nil || target.Active {
		return nil
	}

	confirmationCode := domain.NewConfirmationCode(
		target.ID,
		domain.PurposeAccountConfirmation,
		uc.confirmationTTL,
	)

	if err := uc.gateway.SaveConfirmationCode(ctx, confirmationCode); err != nil {
		return err
	}

	if err := uc.gateway.SendConfirmationEmail(ctx, *target, confirmationCode); err != nil {
		uc.logger.Error("failed to send confirmation email", err, map[string]interface{}{
			logFieldUserID: target.ID.String(),
		})
	}

	return nil
}
