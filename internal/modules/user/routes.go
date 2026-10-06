package user

import (
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/core/module"
)

func (m *Module) Routes() []module.Route {
	public := func(method, path string, h http.HandlerFunc) module.Route {
		return module.Route{
			Method:      method,
			Path:        path,
			Protected:   module.Public,
			RateLimited: module.RateLimited,
			Handler:     h,
		}
	}

	return []module.Route{
		public(http.MethodPost, "/v1/auth/signup", m.handler.SignUp),
		public(http.MethodPost, "/v1/auth/signin", m.handler.SignIn),
		public(http.MethodPost, "/v1/auth/refresh-token", m.handler.RefreshToken),
		public(http.MethodPost, "/v1/auth/confirm-account", m.handler.ConfirmAccount),
		public(
			http.MethodPost,
			"/v1/auth/resend-confirmation-code",
			m.handler.ResendConfirmationCode,
		),
		public(http.MethodPost, "/v1/auth/recovery-code", m.handler.SetRecoveryCode),
		public(
			http.MethodPost,
			"/v1/auth/recovery-code/validate",
			m.handler.ValidateRecoveryCode,
		),
		public(http.MethodPut, "/v1/auth/password", m.handler.UpdatePassword),
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/me",
			Protected:   module.Protected,
			RateLimited: module.RateLimited,
			Handler:     m.handler.Me,
		},
	}
}
