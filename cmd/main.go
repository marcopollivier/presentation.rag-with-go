package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/marcopollivier/rag-go-ex01/internal/handlers"
	"github.com/marcopollivier/rag-go-ex01/internal/openai"
	"github.com/marcopollivier/rag-go-ex01/internal/qdrant"
	"github.com/marcopollivier/rag-go-ex01/internal/rag"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configurar logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel) // Voltando para Info

	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		logger.Warn("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Validar variáveis de ambiente obrigatórias
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY é obrigatória")
	}

	qdrantURL := os.Getenv("QDRANT_URL")
	if qdrantURL == "" {
		qdrantURL = "http://localhost:6333"
	}

	// Inicializar clientes
	logger.Info("Inicializando cliente OpenAI...")
	openaiClient := openai.NewClient(openaiAPIKey, logger)

	logger.Info("Inicializando cliente Qdrant...")
	qdrantClient, err := qdrant.NewClient("localhost", 6333, "rag_documents", logger)
	if err != nil {
		log.Fatalf("Erro ao inicializar cliente Qdrant: %v", err)
	}

	// Aguardar Qdrant estar pronto
	logger.Info("Aguardando Qdrant ficar disponível...")
	for i := 0; i < 30; i++ {
		resp, err := http.Get("http://localhost:6333/")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			logger.Info("Qdrant está disponível!")
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i == 29 {
			log.Fatalf("Timeout aguardando Qdrant ficar disponível")
		}
		logger.Warnf("Tentativa %d/30: Qdrant não disponível, aguardando...", i+1)
		time.Sleep(2 * time.Second)
	}

	// Inicializar serviços
	logger.Info("Inicializando serviço RAG...")
	ragService := rag.NewService(openaiClient, qdrantClient, logger)

	// Inicializar handlers
	handler := handlers.NewHandler(ragService, logger)

	// Configurar Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware de CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Rotas da API
	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.Health)
		api.GET("/stats", handler.GetStats)

		api.POST("/index", handler.IndexDocuments)         // Indexar documentos
		api.POST("/index/sample", handler.IndexSampleData) // Indexar dados de exemplo

		api.POST("/query", handler.Query)     // Query principal
		api.GET("/query", handler.QuickQuery) // Query via GET para testes
	}

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("Iniciando servidor na porta %s", port)
	logger.Info("Endpoints disponíveis:")
	logger.Info("  GET  /                     - Informações da API")
	logger.Info("  GET  /api/v1/health        - Health check")
	logger.Info("  GET  /api/v1/stats         - Estatísticas do sistema")
	logger.Info("  GET  /api/v1/query?q=...   - Query rápida")
	logger.Info("  POST /api/v1/query         - Query principal")
	logger.Info("  POST /api/v1/index         - Indexar documentos")
	logger.Info("  POST /api/v1/index/sample  - Indexar dados de exemplo")

	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
