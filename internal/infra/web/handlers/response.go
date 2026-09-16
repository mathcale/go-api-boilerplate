package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type (
	Response interface {
		Respond(w http.ResponseWriter, statusCode int, data interface{}, headers map[string]string)
		RespondPlainText(w http.ResponseWriter, statusCode int, data string, headers map[string]string)
		RespondWithError(w http.ResponseWriter, err error, headers map[string]string)
	}

	response struct {
		logger      logger.Logger
		environment string
	}
)

func NewResponse(l logger.Logger, environment string) *response {
	return &response{
		logger:      l,
		environment: environment,
	}
}

func (h *response) Respond(
	w http.ResponseWriter,
	statusCode int,
	data interface{},
	headers map[string]string,
) {
	setHeaders(w, headers)
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(&data); err != nil {
			h.logger.Error("failed to encode response body", err, nil)
		}
	}
}

func (h *response) RespondPlainText(
	w http.ResponseWriter,
	statusCode int,
	data string,
	headers map[string]string,
) {
	w.Header().Set("Content-Type", "text/plain")

	for key, value := range headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(data)); err != nil {
		h.logger.Error("failed to write plain text response", err, nil)
	}
}

func (h *response) RespondWithError(w http.ResponseWriter, err error, headers map[string]string) {
	status := http.StatusInternalServerError
	output := dto.ErrorOutput{Message: "internal server error"}

	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		status = kindToStatus(appErr.Kind())
		output.Message = messageForStatus(status)
		output.Code = appErr.BusinessCode()

		if h.environment != config.EnvironmentProduction {
			output.Cause = appErr.Error()
		}
	} else if err != nil && h.environment != config.EnvironmentProduction {
		output.Cause = err.Error()
	}

	if status >= http.StatusInternalServerError {
		h.logger.Error("request failed", err, map[string]interface{}{"status": status})
	}

	setHeaders(w, headers)
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(output); encErr != nil {
		h.logger.Error("failed to encode error response", encErr, nil)
	}
}

func kindToStatus(kind apperror.Kind) int {
	switch kind {
	case apperror.ParseKind, apperror.ValidationKind:
		return http.StatusBadRequest
	case apperror.UnauthorizedKind:
		return http.StatusUnauthorized
	case apperror.ForbiddenKind:
		return http.StatusForbidden
	case apperror.NotFoundKind:
		return http.StatusNotFound
	case apperror.ConflictKind:
		return http.StatusConflict
	case apperror.TooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func messageForStatus(status int) string {
	if msg := http.StatusText(status); msg != "" {
		return msg
	}

	return "internal server error"
}

func setHeaders(w http.ResponseWriter, headers map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")

	for key, value := range headers {
		w.Header().Set(key, value)
	}
}
