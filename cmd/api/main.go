package main

import (
	"context"
	"log"

	"github.com/mathcale/go-api-boilerplate/config"
	_ "github.com/mathcale/go-api-boilerplate/docs"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/di"
)

// @title						Go API Boilerplate
// @version					1.0
// @description				A clean-architecture Go REST API boilerplate with JWT auth.
// @BasePath					/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and the access token.
func main() {
	ctx := context.Background()

	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	deps, err := di.NewDependencyInjector(cfg).Inject(ctx)
	if err != nil {
		log.Fatalf("Failed to inject dependencies: %v", err)
	}

	if err := deps.WebServer.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
