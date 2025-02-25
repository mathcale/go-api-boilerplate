package handlers

import (
	"net/http"
)

type Ping interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

type ping struct {
	response Response
}

func NewPingHandler(r Response) Ping {
	return &ping{
		response: r,
	}
}

func (h *ping) Ping(w http.ResponseWriter, r *http.Request) {
	h.response.RespondPlainText(w, http.StatusOK, "pong", nil)
}
