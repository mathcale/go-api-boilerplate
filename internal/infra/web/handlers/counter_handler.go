package handlers

import (
	"net/http"

	uc "github.com/mathcale/go-api-boilerplate/internal/usecases/counter"
)

type CounterHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type counterHandler struct {
	response       Response
	counterUseCase uc.CounterUseCase
}

func NewCounterHandler(
	r Response,
	uc uc.CounterUseCase,
) CounterHandler {
	return &counterHandler{
		response:       r,
		counterUseCase: uc,
	}
}

func (h *counterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	counter, err := h.counterUseCase.Execute()
	if err != nil {
		h.response.RespondWithError(w, http.StatusInternalServerError, err, nil)
		return
	}

	data := map[string]int{
		"counter": counter,
	}

	h.response.Respond(w, http.StatusOK, data, nil)
}
