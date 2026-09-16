package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type (
	AuthHandler interface {
		SignUp(w http.ResponseWriter, r *http.Request)
		SignIn(w http.ResponseWriter, r *http.Request)
		RefreshToken(w http.ResponseWriter, r *http.Request)
		ConfirmAccount(w http.ResponseWriter, r *http.Request)
		ResendConfirmationCode(w http.ResponseWriter, r *http.Request)
		Me(w http.ResponseWriter, r *http.Request)
		SetRecoveryCode(w http.ResponseWriter, r *http.Request)
		ValidateRecoveryCode(w http.ResponseWriter, r *http.Request)
		UpdatePassword(w http.ResponseWriter, r *http.Request)
	}

	AuthUseCases struct {
		SignUp               auth.SignUpUseCase
		SignIn               auth.SignInUseCase
		RefreshToken         auth.RefreshTokenUseCase
		ConfirmAccount       auth.ConfirmAccountUseCase
		ResendConfirmation   auth.ResendConfirmationCodeUseCase
		Me                   auth.MeUseCase
		SetRecoveryCode      auth.SetRecoveryCodeUseCase
		ValidateRecoveryCode auth.ValidateRecoveryCodeUseCase
		UpdatePassword       auth.UpdatePasswordUseCase
	}

	authHandler struct {
		response Response
		validate *validator.Validate
		useCases AuthUseCases
	}
)

func NewAuthHandler(response Response, useCases AuthUseCases) AuthHandler {
	return &authHandler{
		response: response,
		validate: validator.New(validator.WithRequiredStructEnabled()),
		useCases: useCases,
	}
}

// @Summary		Register a new account
// @Description	Creates an inactive user and emails an account-confirmation code.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body		dto.SignUpInput	true	"Sign-up payload"
// @Success		201		{object}	dto.SignUpOutput
// @Failure		400		{object}	dto.ErrorOutput
// @Failure		409		{object}	dto.ErrorOutput
// @Router			/v1/auth/signup [post]
func (h *authHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var input dto.SignUpInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	created, err := h.useCases.SignUp.Execute(r.Context(), auth.SignUpInput{
		Name:      input.Name,
		Surname:   input.Surname,
		AvatarURL: input.AvatarURL,
		Email:     input.Email,
		Password:  input.Password,
	})
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusCreated, dto.SignUpOutput{
		ID:      created.ID.String(),
		Name:    created.Name,
		Surname: created.Surname,
		Email:   created.Email,
		Active:  created.Active,
	}, nil)
}

// @Summary		Sign in
// @Description	Exchanges credentials for an access + refresh token pair.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body		dto.SignInInput	true	"Credentials"
// @Success		200		{object}	dto.TokenPairOutput
// @Failure		400		{object}	dto.ErrorOutput
// @Failure		401		{object}	dto.ErrorOutput
// @Router			/v1/auth/signin [post]
func (h *authHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var input dto.SignInInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	tokens, err := h.useCases.SignIn.Execute(r.Context(), input.Email, input.Password)
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusOK, dto.TokenPairOutput{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil)
}

// @Summary		Refresh tokens
// @Description	Issues a new access + refresh token pair from a valid refresh token.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body		dto.RefreshTokenInput	true	"Refresh token"
// @Success		200		{object}	dto.TokenPairOutput
// @Failure		401		{object}	dto.ErrorOutput
// @Router			/v1/auth/refresh-token [post]
func (h *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input dto.RefreshTokenInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	tokens, err := h.useCases.RefreshToken.Execute(r.Context(), input.RefreshToken)
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusOK, dto.TokenPairOutput{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil)
}

// @Summary		Confirm account
// @Description	Activates an account using its confirmation code.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body	dto.ConfirmAccountInput	true	"Confirmation payload"
// @Success		204		"No Content"
// @Failure		400		{object}	dto.ErrorOutput
// @Router			/v1/auth/confirm-account [post]
func (h *authHandler) ConfirmAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.ConfirmAccountInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	userID, err := parseUUID(input.UserID, "confirm_account")
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	if err := h.useCases.ConfirmAccount.Execute(r.Context(), userID, input.Code); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusNoContent, nil, nil)
}

// @Summary		Resend confirmation code
// @Description	Emails a fresh account-confirmation code.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body	dto.ResendConfirmationCodeInput	true	"Email payload"
// @Success		204		"No Content"
// @Router			/v1/auth/resend-confirmation-code [post]
func (h *authHandler) ResendConfirmationCode(w http.ResponseWriter, r *http.Request) {
	var input dto.ResendConfirmationCodeInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	if err := h.useCases.ResendConfirmation.Execute(r.Context(), input.Email); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusNoContent, nil, nil)
}

// @Summary		Current user
// @Description	Returns the authenticated user's profile.
// @Tags			auth
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	dto.MeOutput
// @Failure		401	{object}	dto.ErrorOutput
// @Router			/v1/auth/me [get]
func (h *authHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := webcontext.UserIDFromContext(r.Context())
	if !ok {
		unauthorized := apperror.BE0003_UNAUTHORIZED

		h.response.RespondWithError(w, apperror.New(
			nil, "missing authenticated user", apperror.UnauthorizedKind,
			apperror.WebHandlerOrigin, "me", &unauthorized, nil,
		), nil)
		return
	}

	u, err := h.useCases.Me.Execute(r.Context(), userID)
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusOK, dto.MeOutput{
		ID:        u.ID.String(),
		Name:      u.Name,
		Surname:   u.Surname,
		AvatarURL: u.AvatarURL,
		Email:     u.Email,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil)
}

// @Summary		Request password recovery
// @Description	Emails a password-recovery code if the account exists.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body	dto.SetRecoveryCodeInput	true	"Email payload"
// @Success		204		"No Content"
// @Router			/v1/auth/recovery-code [post]
func (h *authHandler) SetRecoveryCode(w http.ResponseWriter, r *http.Request) {
	var input dto.SetRecoveryCodeInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	if err := h.useCases.SetRecoveryCode.Execute(r.Context(), input.Email); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusNoContent, nil, nil)
}

// @Summary		Validate recovery code
// @Description	Checks a recovery code without consuming it.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body	dto.ValidateRecoveryCodeInput	true	"Recovery payload"
// @Success		204		"No Content"
// @Failure		400		{object}	dto.ErrorOutput
// @Router			/v1/auth/recovery-code/validate [post]
func (h *authHandler) ValidateRecoveryCode(w http.ResponseWriter, r *http.Request) {
	var input dto.ValidateRecoveryCodeInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	userID, err := parseUUID(input.UserID, "validate_recovery_code")
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	if err := h.useCases.ValidateRecoveryCode.Execute(r.Context(), userID, input.Code); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusNoContent, nil, nil)
}

// @Summary		Update password
// @Description	Sets a new password using a valid recovery code.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body	dto.UpdatePasswordInput	true	"Update payload"
// @Success		204		"No Content"
// @Failure		400		{object}	dto.ErrorOutput
// @Router			/v1/auth/password [put]
func (h *authHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var input dto.UpdatePasswordInput
	if err := h.decodeAndValidate(r, &input); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	userID, err := parseUUID(input.UserID, "update_password")
	if err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	if err := h.useCases.UpdatePassword.Execute(
		r.Context(), userID, input.Code, input.NewPassword,
	); err != nil {
		h.response.RespondWithError(w, err, nil)
		return
	}

	h.response.Respond(w, http.StatusNoContent, nil, nil)
}

func (h *authHandler) decodeAndValidate(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		code := apperror.BE0001_INVALID_INPUT

		return apperror.New(
			err, "invalid request body", apperror.ParseKind,
			apperror.WebHandlerOrigin, "auth", &code, nil,
		)
	}

	if err := h.validate.Struct(dst); err != nil {
		code := apperror.BE0002_INPUT_VALIDATION_FAILED

		return apperror.New(
			err, "request validation failed", apperror.ValidationKind,
			apperror.WebHandlerOrigin, "auth", &code, nil,
		)
	}

	return nil
}

func parseUUID(raw, origin string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		code := apperror.BE0001_INVALID_INPUT

		return uuid.Nil, apperror.New(
			err, "invalid uuid", apperror.ValidationKind,
			apperror.WebHandlerOrigin, origin, &code, nil,
		)
	}

	return id, nil
}
