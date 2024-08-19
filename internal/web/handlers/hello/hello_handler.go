package hello

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/web/handlers"
)

type HelloHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type helloHandler struct {
	responseHandler handlers.ResponseHandler
}

func NewHelloHandler(rh handlers.ResponseHandler) HelloHandler {
	return &helloHandler{
		responseHandler: rh,
	}
}

func (h *helloHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"message": "Hello, World!",
	}

	h.responseHandler.Respond(w, http.StatusOK, data, nil)
}
