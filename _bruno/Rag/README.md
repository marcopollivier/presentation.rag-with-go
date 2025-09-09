# RAG-Go API - Coleção Bruno

Esta coleção contém todos os endpoints disponíveis para o sistema RAG (Retrieval-Augmented Generation) implementado em Go.

## 🚀 Endpoints Disponíveis

### 1. Healthcheck

- **Método**: GET  
- **URL**: `/api/v1/health`
- **Descrição**: Verifica se o serviço está funcionando
- **Autenticação**: Não requerida

### 2. Consulta Rápida (GET)

- **Método**: GET
- **URL**: `/api/v1/query?q={pergunta}&top_k={numero}&threshold={score}`
- **Descrição**: Consulta RAG via query parameters para testes rápidos
- **Parâmetros**:
  - `q` (obrigatório): A pergunta
  - `top_k` (opcional): Máximo de documentos (default: 5)
  - `threshold` (opcional): Score mínimo (default: 0.7)

### 3. Consulta Completa (POST)

- **Método**: POST
- **URL**: `/api/v1/query`
- **Descrição**: Consulta RAG completa via JSON
- **Body**: JSON com query, top_k e threshold

### 4. Indexar Dados de Exemplo

- **Método**: POST
- **URL**: `/api/v1/index/sample`
- **Descrição**: Indexa arquivos da pasta `./documents`
- **Body**: Nenhum

### 5. Indexar Documentos Personalizados

- **Método**: POST
- **URL**: `/api/v1/index`
- **Descrição**: Indexa documentos fornecidos via JSON
- **Body**: JSON com array de documentos

## 🔧 Configuração

1. Certifique-se de que o servidor está rodando na porta 8080
2. Configure as variáveis de ambiente necessárias (OPENAI_API_KEY)
3. Tenha o Qdrant rodando na porta 6333

## 📝 Como usar

1. Execute primeiro o **Healthcheck** para verificar se tudo está funcionando
2. Use **Indexar Dados de Exemplo** para carregar dados iniciais
3. Teste consultas com **Consulta Rápida** ou **Consulta Completa**
4. Para adicionar seus próprios documentos, use **Indexar Documentos Personalizados**

## 🛠️ Variáveis

A coleção está configurada com:

- `baseUrl`: <http://localhost:8080/api/v1>

Você pode modificar esta variável se o servidor estiver rodando em outra porta ou host.
