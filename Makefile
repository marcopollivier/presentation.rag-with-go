BINARY_NAME=rag-go-app

.PHONY: help build run test clean docker-up docker-down
.DEFAULT_GOAL := help

# Desenvolvimento
clean:
	@rm -rf bin/

deps:
	@go mod tidy

build: clean deps
	@go build -o bin/$(BINARY_NAME) cmd/main.go

# Docker
docker-up:
	@docker compose up -d

docker-down:
	@docker compose down

# Podman
podman-up:
	@podman-compose up -d

podman-down:
	@podman-compose down
