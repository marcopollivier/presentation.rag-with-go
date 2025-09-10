package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/marcopollivier/rag-go-ex01/internal/models"
	"github.com/marcopollivier/rag-go-ex01/internal/rag"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	ragService *rag.Service
	logger     *logrus.Logger
}

// NewHandler cria um novo handler
func NewHandler(ragService *rag.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		ragService: ragService,
		logger:     logger,
	}
}

// Query executa uma consulta RAG
func (h *Handler) Query(c *gin.Context) {
	var req models.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Erro ao fazer bind da requisição")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	response, err := h.ragService.Query(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Erro ao processar query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// IndexDocuments indexa documentos
func (h *Handler) IndexDocuments(c *gin.Context) {
	var req models.IndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Erro ao fazer bind da requisição de indexação")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	response, err := h.ragService.IndexDocuments(c.Request.Context(), req.Documents)
	if err != nil {
		h.logger.WithError(err).Error("Erro ao indexar documentos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// IndexSampleData indexa dados de exemplo
func (h *Handler) IndexSampleData(c *gin.Context) {
	h.logger.Info("Indexando dados de exemplo")

	response, err := h.ragService.IndexTextFiles(c.Request.Context(), "./documents")
	if err != nil {
		h.logger.WithError(err).Error("Erro ao indexar dados de exemplo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QuickQuery permite fazer queries via query parameter para testes rápidos
func (h *Handler) QuickQuery(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'q' é obrigatório"})
		return
	}

	// Parâmetros opcionais
	topK := 5 // O número máximo de documentos relevantes que devem ser retornados
	if topKStr := c.Query("top_k"); topKStr != "" {
		if k, err := strconv.Atoi(topKStr); err == nil {
			topK = k
		}
	}

	threshold := float32(0.7) // O score mínimo de similaridade que um documento deve ter para ser considerado relevante
	if thresholdStr := c.Query("threshold"); thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 32); err == nil {
			threshold = float32(t)
		}
	}

	req := models.QueryRequest{
		Query:     query,
		TopK:      topK,
		Threshold: threshold,
	}

	response, err := h.ragService.Query(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Erro ao processar quick query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAllDocuments retorna todos os documentos indexados
func (h *Handler) GetAllDocuments(c *gin.Context) {
	h.logger.Info("Buscando todos os documentos")

	// Parâmetro opcional para limite
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	documents, err := h.ragService.GetAllDocuments(c.Request.Context(), limit)
	if err != nil {
		h.logger.WithError(err).Error("Erro ao buscar todos os documentos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"count":     len(documents),
	})
}

// GetDocumentsBySource retorna documentos filtrados por fonte
func (h *Handler) GetDocumentsBySource(c *gin.Context) {
	source := c.Param("source")
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'source' é obrigatório"})
		return
	}

	h.logger.Infof("Buscando documentos da fonte: %s", source)

	// Parâmetro opcional para limite
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	documents, err := h.ragService.GetDocumentsBySource(c.Request.Context(), source, limit)
	if err != nil {
		h.logger.WithError(err).Error("Erro ao buscar documentos por fonte")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source":    source,
		"documents": documents,
		"count":     len(documents),
	})
}

// GetCollectionInfo retorna informações sobre a coleção Qdrant
func (h *Handler) GetCollectionInfo(c *gin.Context) {
	h.logger.Info("Obtendo informações da coleção")

	info, err := h.ragService.GetCollectionInfo(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Erro ao obter informações da coleção")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, info)
}
