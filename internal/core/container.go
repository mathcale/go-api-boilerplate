package core

import (
	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/internal/core/bcrypt"
	"github.com/mathcale/go-api-boilerplate/internal/core/config"
	"github.com/mathcale/go-api-boilerplate/internal/core/database"
	"github.com/mathcale/go-api-boilerplate/internal/core/email"
	"github.com/mathcale/go-api-boilerplate/internal/core/email/clients"
	"github.com/mathcale/go-api-boilerplate/internal/core/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/core/logger"
	"github.com/mathcale/go-api-boilerplate/internal/core/module"
	"github.com/mathcale/go-api-boilerplate/internal/core/response"
)

type Container struct {
	Config   *config.Config
	Logger   logger.Logger
	DB       *sqlx.DB
	JWT      jwt.JWTAuth
	Bcrypt   bcrypt.Password
	Response response.Response
	Email    email.Client
}

func Build(cfg *config.Config) (*Container, error) {
	isProd := cfg.Environment == config.EnvironmentProduction

	log := logger.NewLogger(cfg.LogLevel, isProd)

	jwtAuth := jwt.NewJWTAuth(
		log,
		[]byte(cfg.AccessTokenSecret),
		[]byte(cfg.RefreshTokenSecret),
		cfg.AccessTokenLifetimeMinutes,
		cfg.RefreshTokenLifetimeMinutes,
		cfg.TokenIssuer,
		cfg.TokenAudience,
	)

	db, err := database.NewDatabase(
		log,
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
		cfg.DatabasePort,
		cfg.DatabaseMaxOpenConns,
		cfg.DatabaseMaxIdleConns,
		cfg.DatabaseConnMaxLifetimeSecs,
		cfg.DatabaseConnMaxIdleTimeSecs,
	).Connect()
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:   cfg,
		Logger:   log,
		DB:       db,
		JWT:      jwtAuth,
		Bcrypt:   bcrypt.NewPassword(),
		Response: response.NewResponse(log, cfg.Environment),
		Email:    clients.NewFake(log, cfg.MailerSenderName, cfg.MailerSenderEmail),
	}, nil
}

func CollectRoutes(mods ...module.Module) []module.Route {
	routes := []module.Route{}

	for _, mod := range mods {
		routes = append(routes, mod.Routes()...)
	}

	return routes
}
