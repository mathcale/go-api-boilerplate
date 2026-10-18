package di

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database"
	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
	"github.com/mathcale/go-api-boilerplate/internal/infra/email/clients"
	usergateway "github.com/mathcale/go-api-boilerplate/internal/infra/gateways/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/ratelimit"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/middlewares"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/clientip"
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
		closers   []func()
	}

	dependencyInjector struct {
		config *config.Config
	}
)

func (d *Dependencies) Close() {
	for _, closeFn := range d.closers {
		closeFn()
	}
}

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
	emailClient := di.newEmailClient(log)

	// Repositories & gateways
	userRepo := usergateway.NewUserRepository(db)

	userGateway := usergateway.NewGateway(
		userRepo,
		emailClient,
		di.config.AccountConfirmationBaseURL,
		di.config.PasswordRecoveryBaseURL,
	)

	// Rate limiting
	limiters, err := di.newRateLimiters()
	if err != nil {
		return nil, err
	}

	// Use-cases
	signInUseCase, err := auth.NewSignInUseCase(
		log, userGateway, password, jwtAuth, limiters.signIn,
	)
	if err != nil {
		limiters.Close()

		return nil, err
	}

	authUseCases := handlers.AuthUseCases{
		SignUp:               auth.NewSignUpUseCase(log, userGateway, password),
		SignIn:               signInUseCase,
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
	authMiddleware := middlewares.NewAuthMiddleware(jwtAuth, userGateway, rh)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(
		limiters.ip,
		limiters.resolver,
		rh,
	)

	globalMiddlewares := web.NewMiddlewaresResolver(
		securityHeadersMiddleware,
		corsMiddleware,
		middlewares.NewBodyLimitMiddleware(middlewares.MaxRequestBodyBytes),
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
		closers:   []func(){limiters.Close},
	}, nil
}

type rateLimiters struct {
	ip       *ratelimit.KeyedLimiter
	signIn   *ratelimit.KeyedLimiter
	resolver *clientip.Resolver
}

func (r rateLimiters) Close() {
	r.ip.Close()
	r.signIn.Close()
}

func (di *dependencyInjector) newRateLimiters() (rateLimiters, error) {
	trustedProxies, err := clientip.ParseTrustedProxies(di.config.TrustedProxies)
	if err != nil {
		return rateLimiters{}, err
	}

	return rateLimiters{
		ip: ratelimit.NewKeyedLimiter(ratelimit.Config{
			RequestsPerMinute: di.config.RateLimitRequestsPerMinute,
			Burst:             di.config.RateLimitBurst,
		}),
		signIn: ratelimit.NewKeyedLimiter(ratelimit.Config{
			RequestsPerMinute: di.config.SignInRateLimitPerMinute,
			Burst:             di.config.SignInRateLimitBurst,
		}),
		resolver: clientip.NewResolver(trustedProxies),
	}, nil
}

func (di *dependencyInjector) newEmailClient(l logger.Logger) email.Client {
	if di.config.MailerProvider == config.MailerProviderSMTP {
		return clients.NewSMTP(l, clients.SMTPConfig{
			Host:        di.config.SMTPHost,
			Port:        di.config.SMTPPort,
			Username:    di.config.SMTPUsername,
			Password:    di.config.SMTPPassword,
			ImplicitTLS: di.config.SMTPTLSMode == config.SMTPTLSModeImplicit,
			SenderName:  di.config.MailerSenderName,
			SenderEmail: di.config.MailerSenderEmail,
		})
	}

	return clients.NewFake(l, di.config.MailerSenderName, di.config.MailerSenderEmail)
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
