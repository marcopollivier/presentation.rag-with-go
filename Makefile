# Makefile para RAG Go

.PHONY: help setup build run test clean docker-up docker-down index-sample

# Variáveis
BINARY_NAME=rag-go-app
DOCKER_COMPOSE=docker compose

help: ## Mostra esta ajuda
	@echo "RAG Go - Comandos Disponíveis:"
	@echo "=============================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Configura ambiente de desenvolvimento
	@echo "🚀 Configurando ambiente..."
	@./scripts/setup-dev.sh

build: ## Compila a aplicação
	@echo "🔨 Compilando aplicação..."
	@go build -o bin/$(BINARY_NAME) cmd/main.go
	@echo "✅ Binário criado em bin/$(BINARY_NAME)"

run: ## Executa a aplicação localmente
	@echo "🏃 Executando aplicação..."
	@go run cmd/main.go

test: ## Executa testes da API
	@echo "🧪 Testando API..."
	@./scripts/test-api.sh

test-unit: ## Executa testes unitários
	@echo "🧪 Executando testes unitários..."
	@go test ./... -v

clean: ## Limpa ambiente e artefatos
	@echo "🧹 Limpando ambiente..."
	@./scripts/cleanup.sh
	@rm -rf bin/

deps: ## Baixa dependências
	@echo "📦 Baixando dependências..."
	@go mod tidy
	@go mod download

docker-up: ## Inicia todos os serviços com Docker
	@echo "🐳 Iniciando serviços..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Para todos os serviços Docker
	@echo "⏹️  Parando serviços..."
	@$(DOCKER_COMPOSE) down

docker-restart: docker-down docker-up ## Reinicia todos os serviços Docker
	@echo "🔄 Reiniciando serviços..."

docker-logs: ## Mostra logs dos containers
	@echo "📋 Logs dos containers:"
	@$(DOCKER_COMPOSE) logs -f

docker-build: ## Builda imagem Docker da aplicação
	@echo "🔨 Buildando imagem Docker..."
	@docker build -t rag-go-app .

index-sample: ## Indexa dados de exemplo
	@echo "📚 Indexando dados de exemplo..."
	@curl -X POST http://localhost:8080/api/v1/index/sample

query: ## Faz uma query de exemplo
	@echo "🔍 Fazendo query de exemplo..."
	@curl "http://localhost:8080/api/v1/query?q=O%20que%20é%20RAG?"

stats: ## Mostra estatísticas do sistema
	@echo "📊 Estatísticas do sistema:"
	@curl -s http://localhost:8080/api/v1/stats | jq .

health: ## Verifica saúde da aplicação
	@echo "❤️  Verificando saúde:"
	@curl -s http://localhost:8080/api/v1/health | jq .

dev: ## Inicia ambiente completo de desenvolvimento
	@echo "🔧 Iniciando ambiente de desenvolvimento..."
	@make docker-up
	@sleep 10
	@make run

demo: ## Executa demo completa
	@echo "🎯 Executando demo..."
	@make docker-up
	@sleep 15
	@make index-sample
	@sleep 5
	@make query
	@make stats

# Comandos de infraestrutura
infra-up: ## Inicia apenas infraestrutura (Qdrant + LangFlow)
	@echo "🏗️  Iniciando infraestrutura..."
	@$(DOCKER_COMPOSE) up -d qdrant langflow

infra-down: ## Para infraestrutura
	@echo "⏹️  Parando infraestrutura..."
	@$(DOCKER_COMPOSE) stop qdrant langflow

# Comandos de qualidade
lint: ## Executa linting
	@echo "🔍 Executando linting..."
	@golangci-lint run || echo "golangci-lint não instalado"

fmt: ## Formata código
	@echo "✨ Formatando código..."
	@go fmt ./...

mod-verify: ## Verifica módulos
	@echo "🔐 Verificando módulos..."
	@go mod verify

security: ## Verifica vulnerabilidades
	@echo "🛡️  Verificando vulnerabilidades..."
	@govulncheck ./... || echo "govulncheck não instalado"

# Comandos de documentação
docs: ## Gera documentação
	@echo "📚 Documentação disponível em README.md"
	@echo "🌐 APIs em http://localhost:8080/"

# Default target
.DEFAULT_GOAL := help
