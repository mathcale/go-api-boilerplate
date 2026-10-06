package user

import (
	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/gateway"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/internal/adapter"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/internal/repository"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/usecases"
)

type Module struct {
	handler handlers.AuthHandler
	gw      gateway.Gateway
}

func New(c *core.Container) (*Module, error) {
	repo := repository.New(c.DB)
	gw := adapter.New(
		repo,
		c.Email,
		c.Config.AccountConfirmationBaseURL,
		c.Config.PasswordRecoveryBaseURL,
	)

	uc := handlers.AuthUseCases{
		SignUp:               usecases.NewSignUpUseCase(c.Logger, gw, c.Bcrypt),
		SignIn:               usecases.NewSignInUseCase(c.Logger, gw, c.Bcrypt, c.JWT),
		RefreshToken:         usecases.NewRefreshTokenUseCase(c.Logger, c.JWT, gw),
		ConfirmAccount:       usecases.NewConfirmAccountUseCase(c.Logger, gw),
		ResendConfirmation:   usecases.NewResendConfirmationCodeUseCase(c.Logger, gw),
		Me:                   usecases.NewMeUseCase(c.Logger, gw),
		SetRecoveryCode:      usecases.NewSetRecoveryCodeUseCase(c.Logger, gw),
		ValidateRecoveryCode: usecases.NewValidateRecoveryCodeUseCase(c.Logger, gw),
		UpdatePassword:       usecases.NewUpdatePasswordUseCase(c.Logger, gw, c.Bcrypt),
	}

	return &Module{handler: handlers.NewAuthHandler(c.Response, uc), gw: gw}, nil
}

func (m *Module) Name() string {
	return "user"
}

func (m *Module) Gateway() gateway.Gateway {
	return m.gw
}
