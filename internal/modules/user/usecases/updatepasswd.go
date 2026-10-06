package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/core/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
)

type (
	UpdatePasswordUseCase interface {
		Execute(ctx context.Context, userID uuid.UUID, code, newPassword string) error
	}

	updatePasswordUseCase struct {
		logger   logger.Logger
		gateway  gateway.Gateway
		password bcrypt.Password
	}
)

func NewUpdatePasswordUseCase(
	l logger.Logger,
	gw gateway.Gateway,
	pw bcrypt.Password,
) UpdatePasswordUseCase {
	return &updatePasswordUseCase{
		logger:   l,
		gateway:  gw,
		password: pw,
	}
}

func (uc *updatePasswordUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	code, newPassword string,
) error {
	recoveryCode, err := uc.gateway.GetValidConfirmationCode(
		ctx,
		userID,
		code,
		domain.PurposePasswordRecovery,
	)
	if err != nil {
		return err
	}

	if recoveryCode == nil {
		mismatch := apperror.BE1002_RECOVERY_CODE_MISMATCH

		return apperror.New(
			nil, "recovery code is invalid", apperror.ValidationKind,
			apperror.UseCaseOrigin, "update_password", &mismatch, nil,
		)
	}

	if recoveryCode.IsExpired() {
		expired := apperror.BE1007_CONFIRMATION_CODE_EXPIRED

		return apperror.New(
			nil, "recovery code has expired", apperror.ValidationKind,
			apperror.UseCaseOrigin, "update_password", &expired, nil,
		)
	}

	hashed, err := uc.password.Hash(newPassword)
	if err != nil {
		return err
	}

	if err := uc.gateway.UpdatePassword(ctx, userID, *hashed); err != nil {
		return err
	}

	return uc.gateway.MarkConfirmationCodeUsed(ctx, recoveryCode.ID)
}
