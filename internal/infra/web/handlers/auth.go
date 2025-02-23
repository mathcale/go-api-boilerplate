package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	authuc "github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type Auth interface {
	SignIn(w http.ResponseWriter, r *http.Request)
	SignUp(w http.ResponseWriter, r *http.Request)
}

type auth struct {
	response Response
	signUpUC authuc.SignUpUseCase
	signInUC authuc.SignInUseCase
}

func NewAuthHandler(
	r Response,
	signUpUC authuc.SignUpUseCase,
	signInUC authuc.SignInUseCase,
) Auth {
	return &auth{
		response: r,
		signUpUC: signUpUC,
		signInUC: signInUC,
	}
}

func (h *auth) SignIn(w http.ResponseWriter, r *http.Request) {
	var input dto.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.response.RespondWithError(w, http.StatusBadRequest, err, nil)
		return
	}

	if err := input.Validate(); err != nil {
		h.response.RespondWithError(w, http.StatusUnprocessableEntity, err, nil)
		return
	}

	at, rt, err := h.signInUC.Execute(r.Context(), input.ToDomain())
	if err != nil {
		// FIXME: use correct status code by error
		h.response.RespondWithError(w, http.StatusInternalServerError, err, nil)
		return
	}

	h.response.Respond(w, http.StatusOK, dto.NewSignInOutput(*at, *rt), nil)
	return
}

func (h *auth) SignUp(w http.ResponseWriter, r *http.Request) {
	var input dto.SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.response.RespondWithError(w, http.StatusBadRequest, err, nil)
		return
	}

	if err := input.Validate(); err != nil {
		h.response.RespondWithError(w, http.StatusUnprocessableEntity, err, nil)
		return
	}

	if err := h.signUpUC.Execute(r.Context(), input.ToDomain(true)); err != nil {
		// FIXME: use correct status code by error
		h.response.RespondWithError(w, http.StatusInternalServerError, err, nil)
		return
	}

	h.response.Respond(w, http.StatusCreated, nil, nil)
	return
}
