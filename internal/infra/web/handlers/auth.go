package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
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
		e := apperror.New(
			err, "input parse failed", apperror.ParseKind,
			apperror.WebHandlerOrigin, "auth/signin", nil, nil,
		)

		h.response.RespondWithError(w, e, nil)
		return
	}

	if err := input.Validate(); err != nil {
		e := apperror.New(
			err, "input validation failed", apperror.ValidationKind,
			apperror.WebHandlerOrigin, "auth/signin", nil, nil,
		)

		h.response.RespondWithError(w, e, nil)
		return
	}

	at, rt, err := h.signInUC.Execute(r.Context(), input.ToDomain())
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusOK, dto.NewSignInOutput(*at, *rt), nil)
	return
}

func (h *auth) SignUp(w http.ResponseWriter, r *http.Request) {
	var input dto.SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		e := apperror.New(
			err, "input parse failed", apperror.ParseKind,
			apperror.WebHandlerOrigin, "auth/signup", nil, nil,
		)

		h.response.RespondWithError(w, e, nil)
		return
	}

	if err := input.Validate(); err != nil {
		e := apperror.New(
			err, "input validation failed", apperror.ValidationKind,
			apperror.WebHandlerOrigin, "auth/signup", nil, nil,
		)

		h.response.RespondWithError(w, e, nil)
		return
	}

	if err := h.signUpUC.Execute(r.Context(), input.ToDomain(true)); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusCreated, nil, nil)
	return
}
