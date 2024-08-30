package di

import (
	"database/sql"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
	counteruc "github.com/mathcale/go-api-boilerplate/internal/usecases/counter"
	"github.com/mathcale/go-api-boilerplate/internal/web"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers"
	counterhandler "github.com/mathcale/go-api-boilerplate/internal/web/handlers/counter"
	"github.com/mathcale/go-api-boilerplate/internal/web/handlers/hello"
	"github.com/mathcale/go-api-boilerplate/internal/web/middlewares"
)

type DependencyInjector interface {
	Inject() (*Dependencies, error)
}

type dependencyInjector struct {
	config *config.Config
	db     *sql.DB
}

type Dependencies struct {
	WebServer web.Server
}

func NewDependencyInjector(cfg *config.Config, db *sql.DB) DependencyInjector {
	return &dependencyInjector{
		config: cfg,
		db:     db,
	}
}

func (di *dependencyInjector) Inject() (*Dependencies, error) {
	// General
	logger := logger.NewLogger(di.config.LogLevel)
	rh := handlers.NewResponseHandler()

	// Use-cases
	counterUseCase := counteruc.NewCounterUseCase(logger)

	// Middlewares
	loggingMiddleware := middlewares.NewLoggingMiddleware(logger)

	// Handlers
	helloHandler := hello.NewHelloHandler(rh)
	counterHandler := counterhandler.NewCounterHandler(rh, counterUseCase)

	// Web server setup
	handlers := web.NewRouter(helloHandler, counterHandler).Handlers()
	middlewares := web.NewMiddlewaresResolver(loggingMiddleware).Resolve()
	webServer := web.NewServer(logger, di.config.WebServerPort, handlers, middlewares)

	return &Dependencies{
		WebServer: webServer,
	}, nil
}
