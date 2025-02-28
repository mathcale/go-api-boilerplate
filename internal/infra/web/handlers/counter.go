package handlers

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	uc "github.com/mathcale/go-api-boilerplate/internal/usecases/counter"
)

type Counter interface {
	Count(w http.ResponseWriter, r *http.Request)
}

type counter struct {
	response       Response
	counterUseCase uc.CounterUseCase
}

func NewCounterHandler(
	r Response,
	uc uc.CounterUseCase,
) Counter {
	return &counter{
		response:       r,
		counterUseCase: uc,
	}
}

// Count godoc
//
//	@Summary		Increments an in-memory counter
//	@Description	Increments an in-memory counter up to a limit value
//	@Tags			counter
//	@Produce		json
//	@Success		200	{object}	dto.CounterOutput
//	@Failure		500	{object}	dto.ErrorOutput
//	@Router			/v1/counter [get]
func (h *counter) Count(w http.ResponseWriter, r *http.Request) {
	counter, err := h.counterUseCase.Execute(3)
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	data := dto.CounterOutput{
		Counter: counter,
	}

	h.response.Respond(w, http.StatusOK, data, nil)
}
