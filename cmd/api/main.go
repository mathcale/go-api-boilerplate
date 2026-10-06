package main

import (
	"log"

	_ "github.com/mathcale/go-api-boilerplate/docs"
	"github.com/mathcale/go-api-boilerplate/internal/core"
	"github.com/mathcale/go-api-boilerplate/internal/core/config"
	"github.com/mathcale/go-api-boilerplate/internal/core/httpserver"
	"github.com/mathcale/go-api-boilerplate/internal/modules/system"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user"
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
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	c, err := core.Build(cfg)
	if err != nil {
		log.Fatalf("Failed to build container: %v", err)
	}

	userMod, err := user.New(c)
	if err != nil {
		log.Fatalf("Failed to build user module: %v", err)
	}

	server := httpserver.New(c, core.CollectRoutes(system.New(c), userMod))

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
