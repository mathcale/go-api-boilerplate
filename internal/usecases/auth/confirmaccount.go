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
	ConfirmAccountUseCase interface {
		Execute(ctx context.Context, userID uuid.UUID, code string) error
	}

	confirmAccountUseCase struct {
		logger  logger.Logger
		gateway gateway.User
	}
)

func NewConfirmAccountUseCase(l logger.Logger, gw gateway.User) ConfirmAccountUseCase {
	return &confirmAccountUseCase{
		logger:  l,
		gateway: gw,
	}
}

func (uc *confirmAccountUseCase) Execute(ctx context.Context, userID uuid.UUID, code string) error {
	target, err := uc.gateway.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if target == nil {
		notFound := apperror.BE1001_USER_NOT_FOUND

		return apperror.New(
			nil, "user not found", apperror.NotFoundKind,
			apperror.UseCaseOrigin, "confirm_account", &notFound, nil,
		)
	}

	if target.Active {
		already := apperror.BE1008_ACCOUNT_ALREADY_CONFIRMED

		return apperror.New(
			nil, "account already confirmed", apperror.ConflictKind,
			apperror.UseCaseOrigin, "confirm_account", &already, nil,
		)
	}

	confirmationCode, err := uc.gateway.GetValidConfirmationCode(
		ctx,
		userID,
		code,
		user.PurposeAccountConfirmation,
	)
	if err != nil {
		return err
	}

	if confirmationCode == nil {
		invalid := apperror.BE1006_CONFIRMATION_CODE_INVALID

		return apperror.New(
			nil, "confirmation code is invalid", apperror.ValidationKind,
			apperror.UseCaseOrigin, "confirm_account", &invalid, nil,
		)
	}

	if confirmationCode.IsExpired() {
		expired := apperror.BE1007_CONFIRMATION_CODE_EXPIRED

		return apperror.New(
			nil, "confirmation code has expired", apperror.ValidationKind,
			apperror.UseCaseOrigin, "confirm_account", &expired, nil,
		)
	}

	if err := uc.gateway.ActivateUser(ctx, userID); err != nil {
		return err
	}

	return uc.gateway.MarkConfirmationCodeUsed(ctx, confirmationCode.ID)
}
