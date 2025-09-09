BINARY_NAME=rag-go-app

.PHONY: help build run test clean docker-up docker-down docker-down-hard podman-up podman-down podman-down-hard
.DEFAULT_GOAL := help

# Desenvolvimento
clean:
	@rm -rf bin/

deps:
	@go mod tidy

build: clean deps
	@go build -o bin/$(BINARY_NAME) main.go

# Docker
docker-up:
	@docker compose up -d

docker-up-e:
	@docker compose up

docker-down:
	@docker compose down

docker-down-hard:
	@echo "🚨 ATENÇÃO: Isso irá parar e remover TODOS os containers, volumes, redes e imagens do Docker!"
	@echo "Pressione Ctrl+C nos próximos 5 segundos para cancelar..."
	@sleep 5
	@echo "Parando containers do compose..."
	@docker compose down --volumes --remove-orphans || true
	@echo "Removendo todos os containers..."
	@docker container prune -f || true
	@docker stop $$(docker ps -aq) 2>/dev/null || true
	@docker rm $$(docker ps -aq) 2>/dev/null || true
	@echo "Removendo todas as imagens..."
	@docker image prune -af || true
	@docker rmi $$(docker images -aq) 2>/dev/null || true
	@echo "Removendo todos os volumes..."
	@docker volume prune -f || true
	@echo "Removendo todas as redes..."
	@docker network prune -f || true
	@echo "Limpando cache do sistema..."
	@docker system prune -af --volumes || true
	@echo "✅ Limpeza completa do Docker finalizada!"

# Podman
podman-up:
	@podman-compose up -d

podman-up-e:
	@podman-compose up

podman-down:
	@podman-compose down

podman-down-hard:
	@echo "🚨 ATENÇÃO: Isso irá parar e remover TODOS os containers, volumes, redes e imagens do Podman!"
	@echo "Pressione Ctrl+C nos próximos 5 segundos para cancelar..."
	@sleep 5
	@echo "Parando containers do compose..."
	@podman-compose down --volumes --remove-orphans || true
	@echo "Removendo todos os containers..."
	@podman container prune -f || true
	@podman stop $$(podman ps -aq) 2>/dev/null || true
	@podman rm $$(podman ps -aq) 2>/dev/null || true
	@echo "Removendo todas as imagens..."
	@podman image prune -af || true
	@podman rmi $$(podman images -aq) 2>/dev/null || true
	@echo "Removendo todos os volumes..."
	@podman volume prune -f || true
	@echo "Removendo todas as redes..."
	@podman network prune -f || true
	@echo "Limpando cache do sistema..."
	@podman system prune -af --volumes || true
	@echo "✅ Limpeza completa do Podman finalizada!"

# Help
help: ## Mostra esta ajuda
	@echo "🐹 RAG-GO - Comandos Disponíveis:"
	@echo ""
	@echo "📦 Desenvolvimento:"
	@echo "  make clean              - Remove arquivos de build"
	@echo "  make deps               - Instala dependências"
	@echo "  make build              - Compila a aplicação"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-up          - Inicia containers em modo daemon"
	@echo "  make docker-up-e        - Inicia containers em modo interativo"
	@echo "  make docker-down        - Para containers"
	@echo "  make docker-down-hard   - ⚠️  PARA TUDO e limpa completamente Docker"
	@echo ""
	@echo "🫖 Podman:"
	@echo "  make podman-up          - Inicia containers em modo daemon"
	@echo "  make podman-up-e        - Inicia containers em modo interativo"
	@echo "  make podman-down        - Para containers"
	@echo "  make podman-down-hard   - ⚠️  PARA TUDO e limpa completamente Podman"
	@echo ""
	@echo "❓ Ajuda:"
	@echo "  make help               - Mostra esta ajuda"
