package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type Response interface {
	Respond(w http.ResponseWriter, statusCode int, data interface{}, headers map[string]string)
	RespondPlainText(w http.ResponseWriter, statusCode int, data string, headers map[string]string)
	RespondWithError(w http.ResponseWriter, statusCode int, err error, headers map[string]string)
}

type response struct {
	logger logger.Logger
}

func NewResponse(l logger.Logger) *response {
	return &response{
		logger: l,
	}
}

func (h *response) Respond(w http.ResponseWriter, statusCode int, data interface{}, headers map[string]string) {
	setHeaders(w, headers)
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(&data)
	}
}

func (h *response) RespondPlainText(w http.ResponseWriter, statusCode int, data string, headers map[string]string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)

	w.Write([]byte(data))
}

func (h *response) RespondWithError(w http.ResponseWriter, statusCode int, err error, headers map[string]string) {
	setHeaders(w, headers)
	w.WriteHeader(statusCode)

	h.logger.Error(err.Error(), err, nil)

	json.NewEncoder(w).Encode(map[string]string{
		"message": err.Error(),
	})
}

func setHeaders(w http.ResponseWriter, headers map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")

	for key, value := range headers {
		w.Header().Set(key, value)
	}
}
