package di

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database"
	"github.com/mathcale/go-api-boilerplate/internal/infra/email/clients"
	usergateway "github.com/mathcale/go-api-boilerplate/internal/infra/gateways/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

type (
	DependencyInjector interface {
		Inject(ctx context.Context) (*Dependencies, error)
	}

	Dependencies struct {
		Context   context.Context
		WebServer web.Server
	}

	dependencyInjector struct {
		config *config.Config
	}
)

func NewDependencyInjector(cfg *config.Config) DependencyInjector {
	return &dependencyInjector{
		config: cfg,
	}
}

func (di *dependencyInjector) Inject(ctx context.Context) (*Dependencies, error) {
	isProd := di.config.Environment == config.EnvironmentProduction

	// General
	log := logger.NewLogger(di.config.LogLevel, isProd)
	password := bcrypt.NewPassword()

	jwtAuth := jwt.NewJWTAuth(
		log,
		[]byte(di.config.AccessTokenSecret),
		[]byte(di.config.RefreshTokenSecret),
		di.config.AccessTokenLifetimeMinutes,
		di.config.RefreshTokenLifetimeMinutes,
		di.config.TokenIssuer,
		di.config.TokenAudience,
	)

	// Database
	db, err := di.connectToDatabase(log)
	if err != nil {
		return nil, err
	}

	// Email
	emailClient := clients.NewFake(
		log,
		di.config.MailerSenderName,
		di.config.MailerSenderEmail,
	)

	// Repositories & gateways
	userRepo := usergateway.NewUserRepository(db)

	userGateway := usergateway.NewGateway(
		userRepo,
		emailClient,
		di.config.AccountConfirmationBaseURL,
		di.config.PasswordRecoveryBaseURL,
	)

	// Use-cases
	authUseCases := handlers.AuthUseCases{
		SignUp:               auth.NewSignUpUseCase(log, userGateway, password),
		SignIn:               auth.NewSignInUseCase(log, userGateway, password, jwtAuth),
		RefreshToken:         auth.NewRefreshTokenUseCase(log, jwtAuth, userGateway),
		ConfirmAccount:       auth.NewConfirmAccountUseCase(log, userGateway),
		ResendConfirmation:   auth.NewResendConfirmationCodeUseCase(log, userGateway),
		Me:                   auth.NewMeUseCase(log, userGateway),
		SetRecoveryCode:      auth.NewSetRecoveryCodeUseCase(log, userGateway),
		ValidateRecoveryCode: auth.NewValidateRecoveryCodeUseCase(log, userGateway),
		UpdatePassword:       auth.NewUpdatePasswordUseCase(log, userGateway, password),
	}

	// Handlers
	rh := handlers.NewResponse(log, di.config.Environment)

	pingHandler := handlers.NewPingHandler(rh)
	authHandler := handlers.NewAuthHandler(rh, authUseCases)

	// Middlewares
	securityHeadersMiddleware := middlewares.NewSecurityHeadersMiddleware()
	corsMiddleware := middlewares.NewCORSMiddleware(di.corsAllowedOrigins())
	correlationIDMiddleware := middlewares.NewCorrelationIDMiddleware(log)
	loggingMiddleware := middlewares.NewLoggingMiddleware(log)
	authMiddleware := middlewares.NewAuthMiddleware(jwtAuth, rh)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(
		di.config.RateLimitRequestsPerMinute,
		di.config.RateLimitBurst,
		rh,
	)

	globalMiddlewares := web.NewMiddlewaresResolver(
		securityHeadersMiddleware,
		corsMiddleware,
		correlationIDMiddleware,
		loggingMiddleware,
	).Resolve()

	// Web server setup
	routes := web.NewRouter(pingHandler, authHandler).Handlers()

	webServer := web.NewServer(
		log,
		di.config.WebServerPort,
		rh,
		routes,
		globalMiddlewares,
		authMiddleware,
		rateLimitMiddleware,
	)

	return &Dependencies{
		Context:   ctx,
		WebServer: webServer,
	}, nil
}

func (di *dependencyInjector) corsAllowedOrigins() []string {
	raw := strings.TrimSpace(di.config.CORSAllowedOrigins)
	if raw == "" {
		return nil
	}

	origins := strings.Split(raw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return origins
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
