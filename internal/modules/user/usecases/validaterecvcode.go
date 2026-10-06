package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
)

type (
	ValidateRecoveryCodeUseCase interface {
		Execute(ctx context.Context, userID uuid.UUID, code string) error
	}

	validateRecoveryCodeUseCase struct {
		logger  logger.Logger
		gateway gateway.Gateway
	}
)

func NewValidateRecoveryCodeUseCase(
	l logger.Logger,
	gw gateway.Gateway,
) ValidateRecoveryCodeUseCase {
	return &validateRecoveryCodeUseCase{
		logger:  l,
		gateway: gw,
	}
}

func (uc *validateRecoveryCodeUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	code string,
) error {
	recoveryCode, err := uc.gateway.GetValidConfirmationCode(
		ctx, userID, code, domain.PurposePasswordRecovery,
	)
	if err != nil {
		return err
	}

	if recoveryCode == nil {
		mismatch := apperror.BE1002_RECOVERY_CODE_MISMATCH

		return apperror.New(
			nil, "recovery code is invalid", apperror.ValidationKind,
			apperror.UseCaseOrigin, "validate_recovery_code", &mismatch, nil,
		)
	}

	if recoveryCode.IsExpired() {
		expired := apperror.BE1007_CONFIRMATION_CODE_EXPIRED

		return apperror.New(
			nil, "recovery code has expired", apperror.ValidationKind,
			apperror.UseCaseOrigin, "validate_recovery_code", &expired, nil,
		)
	}

	return nil
}
