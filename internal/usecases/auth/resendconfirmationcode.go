package auth

import (
	"context"

	"go-api-boilerplate/internal/domain/gateway"
	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/logger"
)

type (
	ResendConfirmationCodeUseCase interface {
		Execute(ctx context.Context, email string) error
	}

	resendConfirmationCodeUseCase struct {
		logger  logger.Logger
		gateway gateway.User
	}
)

func NewResendConfirmationCodeUseCase(
	l logger.Logger,
	gw gateway.User,
) ResendConfirmationCodeUseCase {
	return &resendConfirmationCodeUseCase{
		logger:  l,
		gateway: gw,
	}
}

func (uc *resendConfirmationCodeUseCase) Execute(ctx context.Context, email string) error {
	target, err := uc.gateway.GetUserByEmailIncludingInactive(ctx, user.NormalizeEmail(email))
	if err != nil {
		return err
	}

	if target == nil || target.Active {
		return nil
	}

	confirmationCode := user.NewConfirmationCode(
		target.ID,
		user.PurposeAccountConfirmation,
		user.AccountConfirmationTTL,
	)

	if err := uc.gateway.SaveConfirmationCode(ctx, confirmationCode); err != nil {
		return err
	}

	if err := uc.gateway.SendConfirmationEmail(ctx, *target, confirmationCode); err != nil {
		uc.logger.Error(
			"failed to send confirmation email", err,
			logFieldUserID, target.ID.String(),
		)
	}

	return nil
}
