package auth

import (
	"context"

	"github.com/google/uuid"

	"go-api-boilerplate/internal/domain/gateway"
	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/apperror"
	"go-api-boilerplate/internal/pkg/bcrypt"
	"go-api-boilerplate/internal/pkg/logger"
)

type (
	UpdatePasswordUseCase interface {
		Execute(ctx context.Context, userID uuid.UUID, code, newPassword string) error
	}

	updatePasswordUseCase struct {
		logger   logger.Logger
		gateway  gateway.User
		password bcrypt.Password
	}
)

func NewUpdatePasswordUseCase(
	l logger.Logger,
	gw gateway.User,
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
		user.PurposePasswordRecovery,
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

	if pwErr := user.ValidatePassword(newPassword); pwErr != nil {
		code := apperror.BE0001_INVALID_INPUT

		return apperror.New(
			pwErr, "invalid new password", apperror.ValidationKind,
			apperror.UseCaseOrigin, "update_password", &code, nil,
		)
	}

	hashed, err := uc.password.Hash(newPassword)
	if err != nil {
		return err
	}

	updated, err := uc.gateway.UpdatePasswordWithCode(ctx, userID, recoveryCode.ID, *hashed)
	if err != nil {
		return err
	}

	if !updated {
		mismatch := apperror.BE1002_RECOVERY_CODE_MISMATCH

		return apperror.New(
			nil, "recovery code is invalid", apperror.ValidationKind,
			apperror.UseCaseOrigin, "update_password", &mismatch, nil,
		)
	}

	return uc.gateway.RevokeUserRefreshTokens(ctx, userID)
}
