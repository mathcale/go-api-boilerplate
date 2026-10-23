package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-api-boilerplate/config"
	_ "go-api-boilerplate/docs"
	"go-api-boilerplate/internal/pkg/di"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	deps, err := di.NewDependencyInjector(cfg).Inject(ctx)
	if err != nil {
		log.Fatalf("Failed to inject dependencies: %v", err)
	}

	serveErr := deps.WebServer.Start(ctx)

	deps.Close()
	stop()

	if serveErr != nil {
		log.Fatalf("Web server stopped: %v", serveErr)
	}
}
