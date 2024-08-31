.PHONY: build run test tidy install-deps setup

build:
	@go build -ldflags="-w -s -buildid=" -trimpath -o ./bin/api ./cmd/api/main.go

run:
	@air -c .air.toml

test:
	@./scripts/test.sh

tidy:
	@go mod tidy

install-deps:
	@go install github.com/air-verse/air@latest
	@go mod tidy

setup: install-deps
