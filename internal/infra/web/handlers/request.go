package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

var requestValidator = newRequestValidator()

func decodeAndValidate(r *http.Request, dst interface{}, origin string) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)
	if err == nil {
		err = ensureSingleJSONValue(decoder)
	}

	if err != nil {
		return bodyError(err, origin)
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

func ensureSingleJSONValue(decoder *json.Decoder) error {
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected data after json body")
		}

		return err
	}

	return nil
}

func bodyError(err error, origin string) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		code := apperror.BE0006_PAYLOAD_TOO_LARGE

		return apperror.New(
			err, "request body too large", apperror.PayloadTooLargeKind,
			apperror.WebHandlerOrigin, origin, &code, nil,
		)
	}

	code := apperror.BE0001_INVALID_INPUT

	return apperror.New(
		err, "invalid request body", apperror.ParseKind,
		apperror.WebHandlerOrigin, origin, &code, nil,
	)
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
