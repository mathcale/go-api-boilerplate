package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

var requestValidator = validator.New(validator.WithRequiredStructEnabled())

func decodeAndValidate(r *http.Request, dst interface{}, origin string) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		code := apperror.BE0001_INVALID_INPUT

		return apperror.New(
			err, "invalid request body", apperror.ParseKind,
			apperror.WebHandlerOrigin, origin, &code, nil,
		)
	}

	if err := requestValidator.Struct(dst); err != nil {
		code := apperror.BE0002_INPUT_VALIDATION_FAILED

		return apperror.New(
			err, "request validation failed", apperror.ValidationKind,
			apperror.WebHandlerOrigin, origin, &code, nil,
		)
	}

	return nil
}

func parseUUID(raw, origin string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		code := apperror.BE0001_INVALID_INPUT

		return uuid.Nil, apperror.New(
			err, "invalid uuid", apperror.ValidationKind,
			apperror.WebHandlerOrigin, origin, &code, nil,
		)
	}

	return id, nil
}
