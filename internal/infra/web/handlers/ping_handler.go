package handlers

import (
	"net/http"
)

type PingHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type helloHandler struct {
	response Response
}

func NewPingHandler(r Response) PingHandler {
	return &helloHandler{
		response: r,
	}
}

func (h *helloHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.response.RespondPlainText(w, http.StatusOK, "pong", nil)
}
