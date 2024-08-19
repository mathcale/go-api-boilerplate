package di

import (
	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
	counteruc "github.com/mathcale/go-api-boilerplate/internal/usecases/counter"
	"github.com/mathcale/go-api-boilerplate/internal/web"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers"
	counterhandler "github.com/mathcale/go-api-boilerplate/internal/web/handlers/counter"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers/hello"
)

type DependencyInjector interface {
	Inject() (*Dependencies, error)
}

type dependencyInjector struct {
	config *config.Config
}

type Dependencies struct {
	WebServer web.Server
}

func NewDependencyInjector(config *config.Config) DependencyInjector {
	return &dependencyInjector{config: config}
}

func (di *dependencyInjector) Inject() (*Dependencies, error) {
	// General
	logger := logger.NewLogger(di.config.LogLevel)
	rh := handlers.NewResponseHandler()

	// Use-cases
	counterUseCase := counteruc.NewCounterUseCase(logger)

	// Handlers
	helloHandler := hello.NewHelloHandler(rh)
	counterHandler := counterhandler.NewCounterHandler(rh, counterUseCase)

	// Web server setup
	handlers := web.NewRouter(helloHandler, counterHandler).Handlers()
	webServer := web.NewServer(logger, handlers, di.config.WebServerPort)

	return &Dependencies{
		WebServer: webServer,
	}, nil
}
