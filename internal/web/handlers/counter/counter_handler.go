package counter

import (
	"net/http"

	uc "github.com/mathcale/go-api-boilerplate/internal/usecases/counter"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers"
)

type CounterHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type counterHandler struct {
	responseHandler handlers.ResponseHandler
	counterUseCase  uc.CounterUseCase
}

func NewCounterHandler(
	rh handlers.ResponseHandler,
	uc uc.CounterUseCase,
) CounterHandler {
	return &counterHandler{
		responseHandler: rh,
		counterUseCase:  uc,
	}
}

func (h *counterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	counter, err := h.counterUseCase.Execute()
	if err != nil {
		h.responseHandler.RespondWithError(w, http.StatusInternalServerError, err, nil)
		return
	}

	data := map[string]int{
		"counter": counter,
	}

	h.responseHandler.Respond(w, http.StatusOK, data, nil)
}
