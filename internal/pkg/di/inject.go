package di

import (
	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
	counteruc "github.com/mathcale/go-api-boilerplate/internal/usecases/counter"
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

func NewDependencyInjector(cfg *config.Config) DependencyInjector {
	return &dependencyInjector{
		config: cfg,
	}
}

func (di *dependencyInjector) Inject() (*Dependencies, error) {
	// General
	logger := logger.NewLogger(di.config.LogLevel)
	rh := handlers.NewResponse()

	// Database
	_, err := di.connectToDatabase(logger)
	if err != nil {
		return nil, err
	}

	// Use-cases
	counterUseCase := counteruc.NewCounterUseCase(logger)

	// Middlewares
	correlationIDMiddleware := middlewares.NewCorrelationIDMiddleware(logger)
	loggingMiddleware := middlewares.NewLoggingMiddleware(logger)

	// Handlers
	pingHandler := handlers.NewPingHandler(rh)
	counterHandler := handlers.NewCounterHandler(rh, counterUseCase)

	// Web server setup
	handlers := web.NewRouter(pingHandler, counterHandler).Handlers()
	middlewares := web.NewMiddlewaresResolver(correlationIDMiddleware, loggingMiddleware).Resolve()
	webServer := web.NewServer(logger, di.config.WebServerPort, handlers, middlewares)

	return &Dependencies{
		WebServer: webServer,
	}, nil
}

func (di *dependencyInjector) connectToDatabase(l logger.Logger) (*sqlx.DB, error) {
	db := database.NewDatabase(
		l,
		di.config.DatabaseHost,
		di.config.DatabaseUser,
		di.config.DatabasePassword,
		di.config.DatabaseName,
		di.config.DatabaseSSLMode,
		di.config.DatabasePort,
		di.config.DatabaseMaxOpenConns,
		di.config.DatabaseMaxIdleConns,
		di.config.DatabaseConnMaxLifetimeSecs,
		di.config.DatabaseConnMaxIdleTimeSecs,
	)

	return db.Connect()
}
