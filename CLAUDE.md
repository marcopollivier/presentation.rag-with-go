# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run locally
go run cmd/main.go

# Build binary
make build               # runs clean → deps → go build -o bin/rag-go-app

# Dependencies
make deps                # go mod tidy

# Docker (recommended)
make docker-up           # start Qdrant + app (detached)
make docker-down         # stop services

# Podman alternative
make podman-up
make podman-down
```

There are no test or lint commands defined in this project.

## Environment

Copy `.env.example` to `.env` and set:
- `OPENAI_API_KEY` (required)
- `QDRANT_URL` (default: `http://localhost:6333`)
- `PORT` (default: `8080`)
- `GIN_MODE` (default: `debug`)

## Architecture

This is a **Retrieval Augmented Generation (RAG)** REST API in Go. The flow:

1. **Index**: Accept documents → generate OpenAI embeddings → store vectors in Qdrant
2. **Query**: Embed the query → vector search in Qdrant → build context → GPT-3.5-turbo generates answer

Layered structure:
```
cmd/main.go              → Gin router setup, dependency injection
internal/handlers/       → HTTP request/response handling, input validation
internal/rag/            → Orchestration (RAG service: ties OpenAI + Qdrant together)
internal/openai/         → OpenAI client (embeddings: ada-002, chat: gpt-3.5-turbo)
internal/qdrant/         → Qdrant HTTP client (collection management, vector CRUD, search)
internal/models/         → Shared structs (Document, QueryRequest, QueryResponse, etc.)
documents/               → Sample text files indexed at startup via /api/v1/index/sample
```

The Qdrant client uses raw `net/http` (not an SDK). Vector size is 1536 (OpenAI ada-002 standard), distance metric is cosine similarity.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/stats` | Collection statistics |
| GET | `/api/v1/query?q=...` | Quick semantic query |
| POST | `/api/v1/query` | Full query with options |
| POST | `/api/v1/index` | Index documents |
| POST | `/api/v1/index/sample` | Index sample data from `documents/` |

Bruno API test files are in `bruno/`.
