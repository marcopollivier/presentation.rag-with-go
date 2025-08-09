# Diagrama de Arquitetura - RAG Go Application

## Arquitetura Geral do Sistema

```mermaid
graph TB
    %% External Services
    User[👤 Usuário]
    OpenAI[🤖 OpenAI API<br/>GPT-3.5 + Embeddings]
    
    %% Main Application
    subgraph "RAG Go Application"
        subgraph "HTTP Layer"
            Router[🌐 Gin Router<br/>Port :8080]
            CORS[🔒 CORS Middleware]
            Router --- CORS
        end
        
        subgraph "API Endpoints"
            Health[📊 /api/v1/health]
            Stats[📈 /api/v1/stats]
            Query[🔍 /api/v1/query]
            Index[📝 /api/v1/index]
            Sample[📄 /api/v1/index/sample]
        end
        
        subgraph "Handler Layer"
            Handler[🎯 Handlers<br/>HTTP Request Processing]
        end
        
        subgraph "Service Layer"
            RAGService[⚙️ RAG Service<br/>Business Logic]
        end
        
        subgraph "Client Layer"
            OpenAIClient[🧠 OpenAI Client<br/>Embeddings & Chat]
            QdrantClient[🗃️ Qdrant Client<br/>Vector Operations]
        end
        
        subgraph "Models"
            Document[📄 Document Model]
            QueryReq[❓ Query Request]
            QueryResp[✅ Query Response]
            IndexReq[📥 Index Request]
            IndexResp[📤 Index Response]
        end
    end
    
    %% Vector Database
    subgraph "Infrastructure"
        Qdrant[(🔍 Qdrant Vector DB<br/>Port :6333)]
        DocStorage[📁 Document Storage<br/>./documents/]
    end
    
    %% User Interactions
    User --> Router
    Router --> Health
    Router --> Stats  
    Router --> Query
    Router --> Index
    Router --> Sample
    
    %% Request Flow
    Health --> Handler
    Stats --> Handler
    Query --> Handler
    Index --> Handler
    Sample --> Handler
    
    Handler --> RAGService
    
    %% Service Dependencies
    RAGService --> OpenAIClient
    RAGService --> QdrantClient
    RAGService --> Document
    RAGService --> QueryReq
    RAGService --> QueryResp
    RAGService --> IndexReq
    RAGService --> IndexResp
    
    %% External Connections
    OpenAIClient -.-> OpenAI
    QdrantClient --> Qdrant
    RAGService --> DocStorage
    
    %% Styling
    classDef userClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef apiClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef serviceClass fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef clientClass fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef dbClass fill:#ffebee,stroke:#c62828,stroke-width:2px
    classDef modelClass fill:#f1f8e9,stroke:#33691e,stroke-width:2px
    
    class User userClass
    class Router,CORS,Health,Stats,Query,Index,Sample apiClass
    class Handler,RAGService serviceClass
    class OpenAIClient,QdrantClient clientClass
    class Qdrant,DocStorage dbClass
    class Document,QueryReq,QueryResp,IndexReq,IndexResp modelClass
```

## Fluxo de Processamento RAG

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant R as 🌐 Router
    participant H as 🎯 Handler
    participant RS as ⚙️ RAG Service
    participant OAI as 🧠 OpenAI Client
    participant QC as 🗃️ Qdrant Client
    participant Q as 🔍 Qdrant DB
    participant OpenAI as 🤖 OpenAI API

    %% Query Flow
    Note over U,OpenAI: Fluxo de Consulta RAG
    U->>+R: POST /api/v1/query
    R->>+H: Query Request
    H->>+RS: Query(context, request)
    
    %% Generate Query Embedding
    RS->>+OAI: GenerateEmbedding(query)
    OAI->>+OpenAI: Create Embedding Request
    OpenAI-->>-OAI: Embedding Vector
    OAI-->>-RS: Query Embedding
    
    %% Search Similar Documents
    RS->>+QC: SearchSimilar(embedding, topK, threshold)
    QC->>+Q: Vector Search Request
    Q-->>-QC: Similar Documents
    QC-->>-RS: Relevant Documents
    
    %% Generate Answer
    RS->>+OAI: GenerateAnswer(query, context)
    OAI->>+OpenAI: Chat Completion Request
    OpenAI-->>-OAI: Generated Answer
    OAI-->>-RS: Final Answer
    
    RS-->>-H: Query Response
    H-->>-R: JSON Response
    R-->>-U: HTTP 200 + Answer

    %% Indexing Flow
    Note over U,OpenAI: Fluxo de Indexação
    U->>+R: POST /api/v1/index
    R->>+H: Index Request
    H->>+RS: IndexDocuments(documents)
    
    loop Para cada documento
        RS->>+OAI: GenerateEmbedding(content)
        OAI->>+OpenAI: Create Embedding Request
        OpenAI-->>-OAI: Content Embedding
        OAI-->>-RS: Document Embedding
        
        RS->>+QC: IndexDocument(document, embedding)
        QC->>+Q: Store Vector + Metadata
        Q-->>-QC: Storage Confirmation
        QC-->>-RS: Index Success
    end
    
    RS-->>-H: Index Response
    H-->>-R: JSON Response
    R-->>-U: HTTP 200 + Stats
```

## Estrutura de Dados

```mermaid
erDiagram
    Document {
        string id
        string content
        map metadata
        string source
        time created
    }
    
    QueryRequest {
        string query
        int top_k
        float32 threshold
    }
    
    QueryResponse {
        string answer
        array relevant_docs
        int64 processing_time_ms
    }
    
    RelevantDocument {
        Document document
        float32 score
    }
    
    IndexRequest {
        array documents
    }
    
    IndexResponse {
        bool success
        int indexed_count
        array failed_docs
        string processing_time
    }
    
    QueryResponse ||--o{ RelevantDocument : contains
    RelevantDocument ||--|| Document : references
    IndexRequest ||--o{ Document : contains
```

## Flowchart Simples do RAG

```mermaid
flowchart LR
    %% Query Flow
    User[� Usuário] --> |"POST /query"| Endpoint[🌐 API Endpoint]
    Endpoint --> App[⚙️ RAG Application]
    App --> |"buscar docs similares"| DB[(�️ Qdrant Vector DB)]
    DB --> |"documentos encontrados"| App
    App --> |"gerar resposta com contexto"| AI[🤖 OpenAI API]
    AI --> |"resposta gerada"| App
    App --> |"JSON response"| Endpoint
    Endpoint --> |"HTTP 200"| User
    
    %% Index Flow
    User2[� Usuário] --> |"POST /index"| Endpoint2[🌐 API Endpoint]
    Endpoint2 --> App2[⚙️ RAG Application]
    App2 --> |"gerar embeddings"| AI2[🤖 OpenAI API]
    AI2 --> |"vetores"| App2
    App2 --> |"armazenar vetores"| DB2[(🗃️ Qdrant Vector DB)]
    DB2 --> |"confirmação"| App2
    App2 --> |"resultado"| Endpoint2
    Endpoint2 --> |"HTTP 200"| User2
    
    %% Styling
    classDef userClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef endpointClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef appClass fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef dbClass fill:#ffebee,stroke:#c62828,stroke-width:2px
    classDef aiClass fill:#fff3e0,stroke:#e65100,stroke-width:2px
    
    class User,User2 userClass
    class Endpoint,Endpoint2 endpointClass
    class App,App2 appClass
    class DB,DB2 dbClass
    class AI,AI2 aiClass
```

## Componentes e Responsabilidades

### 🌐 **HTTP Layer (Gin Router)**

- Roteamento de requisições HTTP
- Middleware de CORS
- Parsing de parâmetros e JSON

### 🎯 **Handler Layer**

- Validação de requests
- Binding de JSON para structs
- Tratamento de erros HTTP
- Logging de requisições

### ⚙️ **RAG Service**

- Lógica de negócio principal
- Orquestração entre OpenAI e Qdrant
- Processamento de documentos
- Geração de respostas contextuais

### 🧠 **OpenAI Client**

- Geração de embeddings (text-embedding-ada-002)
- Geração de respostas (GPT-3.5-turbo)
- Rate limiting e error handling

### 🗃️ **Qdrant Client**

- Operações de busca vetorial
- Indexação de documentos
- Gerenciamento de coleções
- Filtros e threshold de similaridade

### 📄 **Models**

- Estruturas de dados compartilhadas
- Validação de requests
- Serialização JSON

## Endpoints da API

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/` | Informações da API |
| `GET` | `/api/v1/health` | Health check |
| `GET` | `/api/v1/stats` | Estatísticas do sistema |
| `GET` | `/api/v1/query?q=...` | Query rápida via GET |
| `POST` | `/api/v1/query` | Query principal RAG |
| `POST` | `/api/v1/index` | Indexar documentos |
| `POST` | `/api/v1/index/sample` | Indexar dados de exemplo |

## Tecnologias Utilizadas

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Vector Database**: Qdrant
- **AI Service**: OpenAI (GPT-3.5 + Embeddings)
- **Logging**: Logrus
- **Containerization**: Docker & Docker Compose
- **Configuration**: Environment Variables (.env)
