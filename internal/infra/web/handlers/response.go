package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apierror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type Response interface {
	Respond(w http.ResponseWriter, statusCode int, data interface{}, headers map[string]string)
	RespondPlainText(w http.ResponseWriter, statusCode int, data string, headers map[string]string)
	RespondWithError(w http.ResponseWriter, err error, headers map[string]string)
}

type response struct {
	logger logger.Logger
}

func NewResponse(l logger.Logger) Response {
	return &response{
		logger: l,
	}
}

func (h *response) Respond(
	w http.ResponseWriter,
	statusCode int,
	data interface{},
	headers map[string]string,
) {
	h.setHeaders(w, headers)
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(&data)
	}
}

func (h *response) RespondPlainText(
	w http.ResponseWriter,
	statusCode int,
	data string,
	headers map[string]string,
) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)

	w.Write([]byte(data))
}

func (h *response) RespondWithError(
	w http.ResponseWriter,
	err error,
	headers map[string]string,
) {
	var appErr apperror.AppError
	var apiErr apierror.ApiError
	var status int

	if ok := errors.As(err, &appErr); ok {
		status = h.getStatusCodeByErrorKind(appErr)
		apiErr = apierror.New(err, status, appErr.Error(), appErr.BusinessCode())

		h.logger.Error(apiErr.Message(), appErr.OriginalError(), map[string]interface{}{
			"kind":          appErr.Kind(),
			"origin":        appErr.Origin(),
			"origin_name":   appErr.OriginName(),
			"business_code": appErr.BusinessCode(),
			"client_code":   appErr.ClientCode(),
		})
	} else {
		status = http.StatusInternalServerError
		apiErr = apierror.New(err, status, err.Error(), nil)

		h.logger.Error(err.Error(), err, nil)
	}

	h.setHeaders(w, headers)
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]any{
		"message": apiErr.Message(),
		"code":    apiErr.BusinessCode(),
	})
}

func (h *response) setHeaders(w http.ResponseWriter, headers map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")

	for key, value := range headers {
		w.Header().Set(key, value)
	}
}

func (h *response) getStatusCodeByErrorKind(err apperror.AppError) int {
	switch err.Kind() {
	case apperror.ParseKind:
		return http.StatusBadRequest
	case apperror.ValidationKind:
		return http.StatusUnprocessableEntity
	case apperror.RestClientKind:
		return *err.ClientCode()
	case apperror.DatabaseKind:
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	case apperror.BusinessKind, apperror.DependencyKind:
		return http.StatusInternalServerError
	case apperror.ConflictKind:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
