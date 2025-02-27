package fixtures

import (
	"errors"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

func NewUseCaseBusinessError() apperror.AppError {
	return apperror.New(
		errors.New("any-error"), "any-error", apperror.BusinessKind,
		apperror.UseCaseOrigin, "test", nil, nil,
	)
}
