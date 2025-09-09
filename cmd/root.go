package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/marcopollivier/rag-go-ex01/internal/handlers"
	"github.com/marcopollivier/rag-go-ex01/internal/openai"
	"github.com/marcopollivier/rag-go-ex01/internal/qdrant"
	"github.com/marcopollivier/rag-go-ex01/internal/rag"
	"github.com/sirupsen/logrus"
)

func Execute() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel) // Voltando para Info

	if err := godotenv.Load(); err != nil {
		logger.Warn("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// ### OPENAI CLIENT CONFIG ###
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY é obrigatória")
	}
	logger.Info("Inicializando cliente OpenAI...")
	openaiClient := openai.NewClient(openaiAPIKey, logger)

	// ### QDRANT CLIENT CONFIG ###
	logger.Info("Inicializando cliente Qdrant...")
	qdrantClient, err := qdrant.NewClient("localhost", 6333, "rag_documents", logger)
	if err != nil {
		log.Fatalf("Erro ao inicializar cliente Qdrant: %v", err)
	}

	// ### Starting SERVICE ###
	logger.Info("Inicializando serviço RAG...")
	s := rag.NewService(openaiClient, qdrantClient, logger)
	handler := handlers.NewHandler(s, logger)

	router := gin.Default()
	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"service":   "rag-go-ex01",
				"timestamp": "2025-08-09T12:00:00Z",
			})
		})

		api.POST("/index", handler.IndexDocuments)         // Indexar documentos
		api.POST("/index/sample", handler.IndexSampleData) // Indexar dados de exemplo

		api.POST("/query", handler.Query)     // Query principal
		api.GET("/query", handler.QuickQuery) // Query via GET para testes
	}

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
