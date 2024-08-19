package handlers

import (
	"encoding/json"
	"net/http"
)

type ResponseHandler interface {
	Respond(w http.ResponseWriter, statusCode int, data interface{}, headers map[string]string)
	RespondWithError(w http.ResponseWriter, statusCode int, err error, headers map[string]string)
}

type responseHandler struct{}

func NewResponseHandler() *responseHandler {
	return &responseHandler{}
}

func (h *responseHandler) Respond(w http.ResponseWriter, statusCode int, data interface{}, headers map[string]string) {
	setHeaders(w, headers)
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(&data)
	}
}

func (h *responseHandler) RespondWithError(w http.ResponseWriter, statusCode int, err error, headers map[string]string) {
	setHeaders(w, headers)
	w.WriteHeader(statusCode)

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
