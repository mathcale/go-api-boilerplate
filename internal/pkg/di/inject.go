package di

import (
	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/repositories"
	"github.com/mathcale/go-api-boilerplate/internal/infra/gateways"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
	authucs "github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
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
	rh := handlers.NewResponse(logger)
	bcryptPass := bcrypt.NewPassword()
	jwtAuth := jwt.NewJWTAuth(
		logger,
		[]byte(di.config.AccessTokenSecret),
		[]byte(di.config.RefreshTokenSecret),
		di.config.AccessTokenLifetimeMinutes,
		di.config.RefreshTokenLifetimeMinutes,
		di.config.TokenIssuer,
		di.config.TokenAudience,
	)

	// Database
	db, err := di.connectToDatabase(logger)
	if err != nil {
		return nil, err
	}

	// Repositories
	userRepo := repositories.NewUserRepository(db)

	// Gateways START
	// Auth
	signUpGW := gateways.NewSignUpGateway(userRepo)
	signInGW := gateways.NewSignInGateway(userRepo)
	// Gateways END

	// Use-cases START
	// Auth
	signUpUC := authucs.NewSignUpUseCase(logger, bcryptPass, signUpGW)
	signInUC := authucs.NewSignInUseCase(logger, bcryptPass, jwtAuth, signInGW)

	// Counter
	counterUseCase := counteruc.NewCounterUseCase(logger)
	// Use-cases END

	// Middlewares
	correlationIDMiddleware := middlewares.NewCorrelationIDMiddleware(logger)
	loggingMiddleware := middlewares.NewLoggingMiddleware(logger)
	authMiddleware := middlewares.NewAuthMiddleware(logger, jwtAuth)

	// Handlers
	pingHandler := handlers.NewPingHandler(rh)
	authHandler := handlers.NewAuthHandler(rh, signUpUC, signInUC)
	counterHandler := handlers.NewCounterHandler(rh, counterUseCase)

	// Web server setup
	handlers := web.NewRouter(pingHandler, authHandler, counterHandler).Handlers()
	middlewares := web.NewMiddlewaresResolver(correlationIDMiddleware, loggingMiddleware).Resolve()
	webServer := web.NewServer(
		logger,
		di.config.WebServerPort,
		handlers,
		middlewares,
		authMiddleware,
	)

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
