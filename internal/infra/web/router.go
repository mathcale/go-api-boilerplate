package web

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
)

const (
	public       = false
	protected    = true
	notThrottled = false
	rateLimited  = true
)

type (
	Router interface {
		Handlers() []handler
	}

	handler struct {
		path          string
		method        string
		protected     bool
		rateLimited   bool
		requiredRoles []string
		handlerFunc   http.HandlerFunc
	}

	router struct {
		pingHandler handlers.PingHandler
		authHandler handlers.AuthHandler
	}
)

func NewRouter(
	pingHandler handlers.PingHandler,
	authHandler handlers.AuthHandler,
) Router {
	return &router{
		pingHandler: pingHandler,
		authHandler: authHandler,
	}
}

func (r *router) Handlers() []handler {
	return []handler{
		{
			path:        "/ping",
			method:      http.MethodGet,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.pingHandler.Handle,
		},
		{
			path:        "/v1/auth/signup",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.SignUp,
		},
		{
			path:        "/v1/auth/signin",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.SignIn,
		},
		{
			path:        "/v1/auth/refresh-token",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.RefreshToken,
		},
		{
			path:        "/v1/auth/confirm-account",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.ConfirmAccount,
		},
		{
			path:        "/v1/auth/resend-confirmation-code",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.ResendConfirmationCode,
		},
		{
			path:        "/v1/auth/recovery-code",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.SetRecoveryCode,
		},
		{
			path:        "/v1/auth/recovery-code/validate",
			method:      http.MethodPost,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.ValidateRecoveryCode,
		},
		{
			path:        "/v1/auth/password",
			method:      http.MethodPut,
			protected:   public,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.UpdatePassword,
		},
		{
			path:        "/v1/auth/me",
			method:      http.MethodGet,
			protected:   protected,
			rateLimited: rateLimited,
			handlerFunc: r.authHandler.Me,
		},
		{
			// FIXME: Demo of role-based access control where only users whose JWT carries the "admin" role may reach this route
			path:          "/v1/admin/ping",
			method:        http.MethodGet,
			protected:     protected,
			rateLimited:   notThrottled,
			requiredRoles: []string{"admin"},
			handlerFunc:   r.pingHandler.Handle,
		},
	}
}
