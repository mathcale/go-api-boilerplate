.PHONY: build run test tidy run-containers init-database create-migration migrate-up migrate-down install-deps setup clean rename-pkgs all openapi-gen openapi-fmt lint lint-fix fmt

include .env

PLATFORM := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/' -e 's/armv7l/armv7/' -e 's/i386\|i686/386/')

build:
	@go build -ldflags="-w -s -buildid=" -trimpath -o ./bin/api ./cmd/api/main.go

run: run-containers openapi-fmt openapi-gen
	@air -c .air.toml

test:
	@./scripts/test.sh

tidy:
	@go mod tidy

run-containers:
	@docker compose up -d

init-database:
	@./scripts/init-database.sh

create-migration:
	@./scripts/create-migration.sh

migrate-up:
	@migrate -path ./migrations -database "postgresql://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}" -verbose up

migrate-down:
	@migrate -path ./migrations -database "postgresql://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}" -verbose down

openapi-gen:
	@swag init -g cmd/api/main.go -o docs

openapi-fmt:
	@swag fmt

lint:
	@echo "==> Running golangci-lint..."
	@golangci-lint run

lint-fix:
	@echo "==> Running golangci-lint with auto-fixes..."
	@golangci-lint run --fix

fmt:
	@echo "==> Formatting Go code..."
	@golangci-lint fmt

install-deps:
	@if command -v air >/dev/null 2>&1; then \
		echo "==> air already installed, skipping"; \
	else \
		echo "==> Installing air-verse/air"; \
		go install github.com/air-verse/air@latest; \
	fi

	@if command -v migrate >/dev/null 2>&1; then \
		echo "==> migrate already installed, skipping"; \
	else \
		echo "==> Installing golang-migrate/migrate"; \
		curl -L https://github.com/golang-migrate/migrate/releases/download/v4.20.1/migrate.$(PLATFORM)-$(ARCH).tar.gz | tar xvz migrate && sudo mv migrate /usr/local/bin; \
	fi

	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "==> golangci-lint already installed, skipping"; \
	else \
		echo "==> Installing golangci/golangci-lint"; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi

	@if command -v swag >/dev/null 2>&1; then \
		echo "==> swag already installed, skipping"; \
	else \
		echo "==> Installing swaggo/swag"; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi

	@echo "==> Done!"

rename-pkgs:
	@./scripts/rename-pkgs.sh

setup: install-deps run-containers init-database migrate-up
