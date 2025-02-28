.PHONY: build run test tidy run-containers init-database create-migration migrate-up migrate-down install-deps setup clean rename-pkgs all openapi-gen openapi-fmt
include .env

_PLATFORM := $(shell uname -m)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH :=

ifeq ($(_PLATFORM),x86_64)
	ARCH += amd64
endif
ifneq ($(filter %86,$(_PLATFORM)),)
	ARCH += 386
endif
ifneq ($(filter arm%,$(_PLATFORM)),)
	ARCH += arm64
endif

build:
	go build -ldflags="-w -s -buildid=" -trimpath -o ./bin/api ./cmd/api/main.go

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
	@swag init -g cmd/api/main.go

openapi-fmt:
	@swag fmt

install-deps:
	@echo "==> Installing air-verse/air"
	@go install github.com/air-verse/air@latest

	@echo "==> Installing golang-migrate/migrate"
	@curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.$(OS)-$(ARCH).tar.gz | tar xvz migrate && sudo mv migrate /usr/local/bin

	@echo "==> Installing golangci/golangci-lint"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

	@echo "==> Installing swaggo/swag"
	@go install github.com/swaggo/swag/cmd/swag@latest

	@echo "==> Done!"

rename-pkgs:
	@./scripts/rename-pkgs.sh

setup: install-deps run-containers init-database migrate-up
