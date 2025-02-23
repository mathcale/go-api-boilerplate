.PHONY: build run test tidy run-containers init-database create-migration migrate-up migrate-down install-deps setup clean all
include .env

platform := $(shell uname -s | tr '[:upper:]' '[:lower:]')

build:
	@go build -ldflags="-w -s -buildid=" -trimpath -o ./bin/api ./cmd/api/main.go

run: run-containers
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

install-deps:
	@echo "==> Installing air-verse/air"
	@go install github.com/air-verse/air@latest

	@echo "==> Installing golang-migrate/migrate"
	@curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.$(platform)-amd64.tar.gz | tar xvz migrate && sudo mv migrate /usr/local/bin

	@echo "==> Installing golangci/golangci-lint"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5

	@echo "==> Done!"

setup: install-deps run-containers init-database migrate-up
