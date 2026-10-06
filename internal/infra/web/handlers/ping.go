package handlers

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/core/response"
)

type PingHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type helloHandler struct {
	response response.Response
}

func NewPingHandler(r response.Response) PingHandler {
	return &helloHandler{
		response: r,
	}
}

func (h *helloHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.response.RespondPlainText(w, http.StatusOK, "pong", nil)
}
