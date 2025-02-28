package main

import (
	"log"

	"github.com/mathcale/go-api-boilerplate/config"
	_ "github.com/mathcale/go-api-boilerplate/docs"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/di"
)

//	@title			go-api-boilerplate
//	@version		1.0
//	@description	A slightly opinionated HTTP API boilerplate with the Go programming language.

//	@license.name	MIT License
//	@license.url	https://github.com/mathcale/go-api-boilerplate/blob/main/LICENSE

// @externalDocs.description	OpenAPI v2
// @externalDocs.url			https://swagger.io/specification/v2/
func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	inj := di.NewDependencyInjector(cfg)

	deps, err := inj.Inject()
	if err != nil {
		log.Fatalf("Failed to inject dependencies: %v", err)
	}

	if err := deps.WebServer.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
