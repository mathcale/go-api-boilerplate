package main

import (
	"log"

	"github.com/mathcale/go-api-boilerplate/config"
	"github.com/mathcale/go-api-boilerplate/database"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/di"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	conn, err := database.Connect(
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
		cfg.DatabasePort,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	inj := di.NewDependencyInjector(cfg, conn)

	deps, err := inj.Inject()
	if err != nil {
		log.Fatalf("Failed to inject dependencies: %v", err)
	}

	if err := deps.WebServer.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
