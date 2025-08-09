# RAG Go - Retrieval Augmented Generation

Um sistema de RAG (Retrieval Augmented Generation) desenvolvido em Go que combina busca semântica com geração de texto usando OpenAI e Qdrant.

## 🚀 Funcionalidades

- **Busca Semântica**: Utiliza embeddings da OpenAI para busca em documentos
- **Geração de Respostas**: Integração com OpenAI GPT para gerar respostas contextualizadas
- **Banco Vetorial**: Qdrant para armazenamento e busca eficiente de vetores
- **API REST**: Endpoints completos para indexação e consulta
- **Containerização**: Suporte completo ao Docker
- **Logs Estruturados**: Sistema de logging com Logrus

## 🛠️ Tecnologias

- **Go 1.24+** - Linguagem principal
- **Gin** - Framework web
- **Qdrant** - Banco de dados vetorial
- **OpenAI API** - Embeddings e geração de texto
- **Docker** - Containerização
- **Logrus** - Sistema de logs

## 📋 Pré-requisitos

- Go 1.24 ou superior
- Docker e Docker Compose
- Chave da API OpenAI

## ⚙️ Configuração

1. **Clone o repositório:**
```bash
git clone <url-do-repositorio>
cd rag-go-ex01
```

2. **Configure as variáveis de ambiente:**
```bash
# Copie o arquivo de exemplo
cp .env.example .env

# Edite o .env e adicione sua chave da OpenAI
OPENAI_API_KEY=sua_chave_openai_aqui
QDRANT_URL=http://localhost:6333
```

## 🚀 Como Executar

### Opção 1: Docker Compose (Recomendado)

```bash
# Inicia todos os serviços (Qdrant + API)
docker compose up -d

# Verificar se os serviços estão rodando
docker compose ps
```

### Opção 2: Local (apenas a API)

```bash
# 1. Inicie apenas o Qdrant
docker compose up -d qdrant

# 2. Execute a aplicação Go
go run cmd/main.go
```

### Opção 3: Usando Makefile

```bash
# Ver comandos disponíveis
make help

# Executar localmente
make run

# Executar com Docker
make docker-up

# Parar serviços
make docker-down
```

## 📚 Exemplos de Uso da API

### 1. Health Check
```bash
curl http://localhost:8080/api/v1/health
```

**Resposta:**
```json
{
  "service": "rag-go-ex01",
  "status": "ok",
  "timestamp": "2025-08-09T12:00:00Z"
}
```

### 2. Informações da API
```bash
curl http://localhost:8080/
```

**Resposta:**
```json
{
  "message": "RAG Go API - Retrieval Augmented Generation",
  "version": "1.0.0",
  "endpoints": {
    "health": "GET /api/v1/health",
    "stats": "GET /api/v1/stats",
    "quick_query": "GET /api/v1/query?q=sua_pergunta",
    "query": "POST /api/v1/query",
    "index": "POST /api/v1/index",
    "index_sample": "POST /api/v1/index/sample"
  }
}
```

### 3. Estatísticas do Sistema
```bash
curl http://localhost:8080/api/v1/stats
```

**Resposta:**
```json
{
  "stats": {
    "distance": "Cosine",
    "points_count": 2,
    "status": "green",
    "vector_size": 1536
  }
}
```

### 4. Indexar Documentos de Exemplo
```bash
curl -X POST http://localhost:8080/api/v1/index/sample
```

**Resposta:**
```json
{
  "success": true,
  "indexed_count": 2,
  "processing_time": "932.806042ms"
}
```

### 5. Consulta Rápida (GET)
```bash
curl "http://localhost:8080/api/v1/query?q=O%20que%20é%20Go?"
```

**Resposta:**
```json
{
  "answer": "Go é uma linguagem de programação desenvolvida pelo Google conhecida por sua simplicidade, performance e excelente suporte a concorrência através de goroutines e channels...",
  "relevant_docs": [
    {
      "document": {
        "id": "7c01de06-c5a4-47b8-93af-cdc0c05628ac",
        "content": "Go é uma linguagem de programação desenvolvida pelo Google...",
        "metadata": {
          "file_path": "documents/golang_intro.txt",
          "file_size": "706",
          "language": "portuguese"
        },
        "source": "golang_intro.txt",
        "created": "2025-08-09T08:07:07Z"
      },
      "score": 0.86825037
    }
  ],
  "processing_time_ms": 1888
}
```

### 6. Consulta Completa (POST)
```bash
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "O que é RAG?",
    "top_k": 3,
    "threshold": 0.7
  }'
```

**Resposta:**
```json
{
  "answer": "RAG (Retrieval Augmented Generation) é uma técnica avançada de inteligência artificial que combina busca de informações com geração de texto...",
  "relevant_docs": [
    {
      "document": {
        "id": "274c9298-c646-4637-90aa-ba6a4516eca8",
        "content": "RAG (Retrieval Augmented Generation) é uma técnica avançada...",
        "metadata": {
          "file_path": "documents/rag_explanation.txt",
          "file_size": "849",
          "language": "portuguese"
        },
        "source": "rag_explanation.txt",
        "created": "2025-08-09T08:07:07Z"
      },
      "score": 0.95
    }
  ],
  "processing_time_ms": 1542
}
```

### 7. Indexar Documentos Customizados
```bash
curl -X POST http://localhost:8080/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{
    "documents": [
      {
        "content": "Kubernetes é uma plataforma de orquestração de contêineres.",
        "source": "kubernetes_intro.txt",
        "metadata": {
          "category": "devops",
          "language": "portuguese"
        }
      }
    ]
  }'
```

## 🏗️ Estrutura do Projeto

```
.
├── cmd/
│   └── main.go              # Ponto de entrada da aplicação
├── internal/
│   ├── handlers/            # Handlers HTTP
│   ├── models/              # Modelos de dados
│   ├── openai/              # Cliente OpenAI
│   ├── qdrant/              # Cliente Qdrant
│   └── rag/                 # Serviço RAG principal
├── documents/               # Documentos de exemplo
├── docker-compose.yml       # Configuração Docker
├── Dockerfile              # Imagem da aplicação
├── Makefile                # Comandos úteis
└── README.md               # Este arquivo
```

## 🔧 Configurações Avançadas

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `OPENAI_API_KEY` | Chave da API OpenAI | *obrigatório* |
| `QDRANT_URL` | URL do Qdrant | `http://localhost:6333` |
| `PORT` | Porta da API | `8080` |
| `GIN_MODE` | Modo do Gin | `debug` |

### Parâmetros de Query

| Parâmetro | Tipo | Descrição | Padrão |
|-----------|------|-----------|---------|
| `query` | string | Pergunta a ser respondida | *obrigatório* |
| `top_k` | int | Número máximo de documentos | `5` |
| `threshold` | float | Limite mínimo de similaridade | `0.7` |

## 🐳 Serviços Docker

- **rag-go-app**: API principal (porta 8080)
- **qdrant**: Banco vetorial (portas 6333, 6334)
- **langflow**: Interface para fluxos (porta 7860)

## 📝 Logs

A aplicação utiliza logs estruturados em JSON. Para visualizar:

```bash
# Ver logs da aplicação
docker compose logs -f rag-go-app

# Ver logs do Qdrant
docker compose logs -f qdrant
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 📞 Suporte

Para dúvidas ou suporte:
- Abra uma issue no GitHub
- Verifique os logs da aplicação
- Consulte a documentação da OpenAI e Qdrant
