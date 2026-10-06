package di

import (
	"context"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/config"
	"github.com/mathcale/go-api-boilerplate/internal/core/httpserver"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	usergateway "github.com/mathcale/go-api-boilerplate/internal/infra/gateways/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/modules/system"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type (
	DependencyInjector interface {
		Inject(ctx context.Context) (*Dependencies, error)
	}

	Dependencies struct {
		Context   context.Context
		WebServer httpserver.Server
	}

	dependencyInjector struct {
		config *config.Config
	}

	legacyUserModule struct {
		handler handlers.AuthHandler
	}
)

func NewDependencyInjector(cfg *config.Config) DependencyInjector {
	return &dependencyInjector{
		config: cfg,
	}
}

func (di *dependencyInjector) Inject(ctx context.Context) (*Dependencies, error) {
	c, err := core.Build(di.config)
	if err != nil {
		return nil, err
	}

	userRepo := usergateway.NewUserRepository(c.DB)

	userGateway := usergateway.NewGateway(
		userRepo,
		c.Email,
		di.config.AccountConfirmationBaseURL,
		di.config.PasswordRecoveryBaseURL,
	)

	authUseCases := handlers.AuthUseCases{
		SignUp:               auth.NewSignUpUseCase(c.Logger, userGateway, c.Bcrypt),
		SignIn:               auth.NewSignInUseCase(c.Logger, userGateway, c.Bcrypt, c.JWT),
		RefreshToken:         auth.NewRefreshTokenUseCase(c.Logger, c.JWT, userGateway),
		ConfirmAccount:       auth.NewConfirmAccountUseCase(c.Logger, userGateway),
		ResendConfirmation:   auth.NewResendConfirmationCodeUseCase(c.Logger, userGateway),
		Me:                   auth.NewMeUseCase(c.Logger, userGateway),
		SetRecoveryCode:      auth.NewSetRecoveryCodeUseCase(c.Logger, userGateway),
		ValidateRecoveryCode: auth.NewValidateRecoveryCodeUseCase(c.Logger, userGateway),
		UpdatePassword:       auth.NewUpdatePasswordUseCase(c.Logger, userGateway, c.Bcrypt),
	}

	legacy := &legacyUserModule{handler: handlers.NewAuthHandler(c.Response, authUseCases)}

	webServer := httpserver.New(c, core.CollectRoutes(system.New(c), legacy))

	return &Dependencies{
		Context:   ctx,
		WebServer: webServer,
	}, nil
}

func (m *legacyUserModule) Name() string {
	return "user"
}

func (m *legacyUserModule) Routes() []module.Route {
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
