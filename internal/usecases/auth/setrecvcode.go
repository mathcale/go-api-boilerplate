package auth

import (
	"context"

	"go-api-boilerplate/internal/domain/gateway"
	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/logger"
)

type (
	SetRecoveryCodeUseCase interface {
		Execute(ctx context.Context, email string) error
	}

	setRecoveryCodeUseCase struct {
		logger  logger.Logger
		gateway gateway.User
	}
)

func NewSetRecoveryCodeUseCase(l logger.Logger, gw gateway.User) SetRecoveryCodeUseCase {
	return &setRecoveryCodeUseCase{
		logger:  l,
		gateway: gw,
	}
}

func (uc *setRecoveryCodeUseCase) Execute(ctx context.Context, email string) error {
	target, err := uc.gateway.GetUserByEmailIncludingInactive(ctx, user.NormalizeEmail(email))
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	recoveryCode := user.NewConfirmationCode(
		target.ID,
		user.PurposePasswordRecovery,
		user.PasswordRecoveryTTL,
	)

	if err := uc.gateway.SaveConfirmationCode(ctx, recoveryCode); err != nil {
		return err
	}

	if err := uc.gateway.SendRecoveryEmail(ctx, *target, recoveryCode); err != nil {
		uc.logger.Error("failed to send recovery email", err, logFieldUserID, target.ID.String())
	}

	return nil
}
